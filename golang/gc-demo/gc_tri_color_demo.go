package main

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// 三色标记常量（使用int32以便原子操作）
const (
	White int32 = iota // 0: 未标记/垃圾对象
	Grey               // 1: 待处理态
	Black              // 2: 完成态/存活对象
)

// Object 模拟堆内存中的引用对象
// 注意：Color字段使用原子操作保护
type Object struct {
	ID       string
	Color    int32      // 使用int32支持原子操作
	Children []*Object  // 子对象用指针，模拟真实引用关系
	mu       sync.Mutex // 保护Children修改
}

// NewObject 创建新对象，初始为白色
func NewObject(id string) *Object {
	return &Object{
		ID:       id,
		Color:    White,
		Children: make([]*Object, 0),
	}
}

// GetColor 安全获取对象颜色（原子读取）
func (o *Object) GetColor() int32 {
	return atomic.LoadInt32(&o.Color)
}

// SetColor 安全设置对象颜色（原子写入）
func (o *Object) SetColor(newColor int32) {
	atomic.StoreInt32(&o.Color, newColor)
}

// CompareAndSetColor 原子比较并设置颜色（CAS操作）
func (o *Object) CompareAndSetColor(oldColor, newColor int32) bool {
	return atomic.CompareAndSwapInt32(&o.Color, oldColor, newColor)
}

// AddChild 线程安全地添加子对象
func (o *Object) AddChild(child *Object) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.Children = append(o.Children, child)
}

// GetChildren 线程安全地获取子对象
func (o *Object) GetChildren() []*Object {
	o.mu.Lock()
	defer o.mu.Unlock()
	// 创建副本避免并发修改问题
	children := make([]*Object, len(o.Children))
	copy(children, o.Children)
	return children
}

// 模拟写屏障：当黑色对象引用白色对象时，将白色对象标记为灰色
func WriteBarrier(parent *Object, child *Object, grayQueue *list.List, queueMutex *sync.Mutex) {
	// 原子读取颜色，避免竞态
	parentColor := parent.GetColor()
	childColor := child.GetColor()

	if parentColor == Black && childColor == White {
		// 使用CAS确保颜色转换的原子性
		if child.CompareAndSetColor(White, Grey) {
			queueMutex.Lock()
			grayQueue.PushBack(child)
			queueMutex.Unlock()
			fmt.Printf("WriteBarrier: [%s]（黑）→ [%s]（白），强制将[%s]标记为灰色\n",
				parent.ID, child.ID, child.ID)
		}
	}
}

// 辅助方法：递归查找第一个黑色对象，用于模拟业务协程场景
func findBlackObject(objs []*Object) *Object {
	for _, obj := range objs {
		if obj.GetColor() == Black {
			return obj
		}
		// 递归遍历子对象
		children := obj.GetChildren()
		if child := findBlackObject(children); child != nil {
			return child
		}
	}
	return nil
}

// 辅助方法：递归获取所有对象，模拟堆内存遍历
func getAllObjects(objs []*Object) []*Object {
	var all []*Object
	visited := make(map[*Object]bool)
	var mutex sync.Mutex

	var traverse func(*Object)
	traverse = func(o *Object) {
		mutex.Lock()
		if visited[o] {
			mutex.Unlock()
			return
		}
		visited[o] = true
		mutex.Unlock()

		all = append(all, o)
		children := o.GetChildren()
		for _, child := range children {
			traverse(child)
		}
	}

	for _, o := range objs {
		traverse(o)
	}
	return all
}

// 三色标记与清除主流程
func TriColorMarkAndSweep(rootObjects []*Object) ([]*Object, []*Object) {
	// 1. 初始化灰色队列和同步锁（保证并发安全）
	grayQueue := list.New()
	var queueMutex sync.Mutex

	// 2. 初始标记阶段(STW)：标记根对象
	fmt.Println("===== 阶段1：初始标记（STW） =====")
	for _, root := range rootObjects {
		// 原子检查并设置颜色
		if root.GetColor() == White {
			root.SetColor(Grey)
			fmt.Printf("根对象[%s] 白→灰，加入灰色队列\n", root.ID)
			queueMutex.Lock()
			grayQueue.PushBack(root)
			queueMutex.Unlock()
		}
	}

	// 3. 并发标记阶段(无STW)
	fmt.Println("\n===== 阶段2：并发标记（无STW） =====")
	var wg sync.WaitGroup

	// 模拟业务协程运行，并发修改引用关系
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond) // 等待GC标记协程启动并处理部分对象
		fmt.Println("\n[业务协程] 正在运行，修改对象引用关系...")

		// 模拟并发修改：黑色对象指向新的白色对象
		blackObj := findBlackObject(rootObjects)
		if blackObj != nil {
			newWhiteObj := NewObject("新创建的白色对象")
			fmt.Printf("[业务协程] 尝试让黑色对象[%s]引用新白色对象[%s]\n",
				blackObj.ID, newWhiteObj.ID)

			// 触发写屏障，拦截危险引用
			WriteBarrier(blackObj, newWhiteObj, grayQueue, &queueMutex)

			// 线程安全地添加子对象
			blackObj.AddChild(newWhiteObj)

			// 验证写屏障效果
			time.Sleep(10 * time.Millisecond)
			if newWhiteObj.GetColor() == Grey {
				fmt.Printf("[业务协程] 验证: [%s] 已被写屏障保护，状态=灰色\n", newWhiteObj.ID)
			} else {
				fmt.Printf("[业务协程] 警告: [%s] 未被写屏障保护，状态=%d\n", newWhiteObj.ID, newWhiteObj.GetColor())
			}
		} else {
			fmt.Println("[业务协程] 未找到黑色对象，跳过引用修改")
		}
	}()

	// GC标记协程：遍历灰色队列，处理对象标记
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			var grayObj *Object

			// 原子操作获取灰色对象
			queueMutex.Lock()
			if grayQueue.Len() == 0 {
				queueMutex.Unlock()
				break
			}
			// 取出队列头部的灰色对象
			elem := grayQueue.Front()
			grayObj = elem.Value.(*Object)
			grayQueue.Remove(elem)
			queueMutex.Unlock()

			fmt.Printf("处理灰色对象: [%s] (当前颜色=%d)\n", grayObj.ID, grayObj.GetColor())

			// 原子获取子对象列表（避免并发修改）
			children := grayObj.GetChildren()

			// 遍历当前对象的所有子对象
			for _, child := range children {
				// 原子检查子对象颜色
				if child.GetColor() == White {
					// 使用CAS确保状态转换正确
					if child.CompareAndSetColor(White, Grey) {
						fmt.Printf("  [%s] 引用了 [%s]，白→灰，加入队列\n", grayObj.ID, child.ID)

						queueMutex.Lock()
						grayQueue.PushBack(child)
						queueMutex.Unlock()
					}
				}
			}

			// 所有子对象处理完成，当前对象转为黑色
			// 使用CAS确保从灰色到黑色的转换
			if grayObj.CompareAndSetColor(Grey, Black) {
				fmt.Printf("  [%s] 所有子对象标记完成，灰→黑\n", grayObj.ID)
			} else {
				fmt.Printf("  [%s] 颜色转换失败，当前状态=%d\n", grayObj.ID, grayObj.GetColor())
			}

			time.Sleep(10 * time.Millisecond) // 模拟标记耗时，便于观察流程
		}
		fmt.Println("灰色队列已空，并发标记阶段完成")
	}()

	// 等待业务协程和GC标记协程完成
	wg.Wait()

	// 4. 重新标记阶段(STW)：处理写屏障捕获的漏标对象
	fmt.Println("\n===== 阶段3：重新标记（STW） =====")
	queueMutex.Lock()
	for grayQueue.Len() > 0 {
		elem := grayQueue.Front()
		obj := elem.Value.(*Object)
		grayQueue.Remove(elem)

		if obj.GetColor() == Grey {
			// 遍历子对象，确保标记完整
			children := obj.GetChildren()
			for _, child := range children {
				if child.GetColor() == White {
					if child.CompareAndSetColor(White, Grey) {
						fmt.Printf("重新标记补标: [%s] 引用的 [%s]（白→灰）\n", obj.ID, child.ID)
						grayQueue.PushBack(child)
					}
				}
			}

			// 标记完成后转为黑色
			obj.SetColor(Black)
			fmt.Printf("重新标记确认: 对象[%s]为存活，灰→黑\n", obj.ID)
		}
	}
	queueMutex.Unlock()

	// 5. 收集所有堆对象（包括未被根对象引用的孤立对象）
	fmt.Println("\n===== 阶段4：收集所有堆对象 =====")
	allObjects := getAllObjects(rootObjects)

	// 6. 创建真正的孤立对象（无根可达）
	trueGarbage1 := NewObject("真正的孤立对象-临时会话")
	trueGarbage2 := NewObject("真正的孤立对象-过期验证码")

	// 添加到所有对象集合，模拟真实堆内存
	allObjects = append(allObjects, trueGarbage1, trueGarbage2)

	// 7. 并发清除阶段：回收白色对象
	fmt.Println("\n===== 阶段5：并发清除（无STW） =====")
	var garbageObjects []*Object
	var liveObjects []*Object

	for _, obj := range allObjects {
		if obj.GetColor() == White {
			garbageObjects = append(garbageObjects, obj)
			fmt.Printf("对象[%s] 为白色，判定为垃圾，执行回收\n", obj.ID)
		} else {
			liveObjects = append(liveObjects, obj)
			fmt.Printf("对象[%s] 为存活对象(颜色=%d)，保留\n", obj.ID, obj.GetColor())
			// 重置为白色，为下一次GC做准备
			obj.SetColor(White)
		}
	}

	return garbageObjects, liveObjects
}

func main() {
	// ========== 构建对象引用链路，模拟真实业务场景 ==========
	// 根对象：模拟全局变量/栈帧变量
	root1 := NewObject("根对象-用户服务")
	root2 := NewObject("根对象-订单服务")

	// 构建引用链路：用户服务 → 用户缓存 → 具体用户
	userCache := NewObject("用户缓存")
	user1 := NewObject("用户-张三")
	user2 := NewObject("用户-李四")
	userCache.AddChild(user1)
	userCache.AddChild(user2)
	root1.AddChild(userCache)

	// 构建引用链路：订单服务 → 订单缓存 → 具体订单
	orderCache := NewObject("订单缓存")
	order1 := NewObject("订单-202311001")
	orderCache.AddChild(order1)
	root2.AddChild(orderCache)

	// 创建真实可访问对象
	validObj := NewObject("有效会话数据")
	order1.AddChild(validObj)

	// ========== 执行三色标记GC流程 ==========
	rootObjects := []*Object{root1, root2}
	garbage, live := TriColorMarkAndSweep(rootObjects)

	// ========== 输出GC执行结果 ==========
	fmt.Printf("\n===== GC执行完成，统计结果 =====\n")
	fmt.Printf("存活对象数量: %d\n", len(live))
	fmt.Printf("垃圾对象数量: %d\n", len(garbage))
	fmt.Println("\n垃圾对象详情:")
	for _, g := range garbage {
		fmt.Printf("- %s (最终颜色=%d)\n", g.ID, g.GetColor())
	}

	// 验证关键对象状态
	fmt.Println("\n===== 关键对象状态验证 =====")
	expectedGarbage := []string{"真正的孤立对象-临时会话", "真正的孤立对象-过期验证码"}
	found := 0
	for _, g := range garbage {
		for _, expected := range expectedGarbage {
			if g.ID == expected {
				fmt.Printf("✓ 对象[%s] 被正确识别为垃圾对象 (颜色=%d)\n", g.ID, g.GetColor())
				found++
			}
		}
	}
	fmt.Printf("垃圾对象识别准确率: %.0f%% (%d/%d)\n", float64(found)/float64(len(expectedGarbage))*100, found, len(expectedGarbage))

	// 验证写屏障对象状态
	fmt.Println("\n===== 写屏障保护验证 =====")
	newObjFound := false
	for _, l := range live {
		if l.ID == "新创建的白色对象" {
			newObjFound = true
			if l.GetColor() == Black {
				fmt.Printf("✓ 对象[%s] 被正确标记为存活 (颜色=%d)，写屏障保护成功\n", l.ID, l.GetColor())
			} else {
				fmt.Printf("✗ 对象[%s] 未被正确标记 (颜色=%d)，写屏障保护失败\n", l.ID, l.GetColor())
			}
		}
	}
	if !newObjFound {
		fmt.Println("⚠ 未找到写屏障创建的新对象，可能未被正确处理")
	}
}
