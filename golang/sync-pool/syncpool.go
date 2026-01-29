package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

/**
sync.Pool是一个临时对象池，用于存储和复用临时对象，以减少内存分配和GC的压力

原理：Pool 中有：本地池 和 受害者缓存
每个p 都有一个本地池（分private 和 shared） private 是 p 私有的 shared是双向链表 可以被多个p 共享

p.Get() ->private -为空->shared拿->为空-> 尝试从其他p偷 -> 为空->new一个
p.Put()->private->为满->shared

在每次GC之前，会将当前Pool中的对象移动到受害者缓存，同时将之前的受害者缓存清空，
这样在GC后，如果还需要对象，可以从victim cache中获取，这样可以将对象的生命周期延长到下一次GC，减少GC压力。

*/

// 跟踪对象生命周期的演示
type TrackedObject struct {
	ID        int
	CreatedAt time.Time
	UsedCount int
}

var objectCounter int
var mu sync.Mutex

func NewTrackedObject() interface{} {
	mu.Lock()
	defer mu.Unlock()

	objectCounter++
	obj := &TrackedObject{
		ID:        objectCounter,
		CreatedAt: time.Now(),
		UsedCount: 1,
	}
	fmt.Printf("🎯 创建新对象 #%d\n", obj.ID)
	return obj
}

func demonstrateVictimCache() {
	pool := &sync.Pool{
		New: NewTrackedObject,
	}

	fmt.Println("\n=== 第1轮：初始状态 ===")

	// 创建3个对象放入pool
	for i := 0; i < 3; i++ {
		obj := pool.Get().(*TrackedObject)
		obj.UsedCount++
		fmt.Printf("获取对象 #%d (使用次数: %d)\n", obj.ID, obj.UsedCount)
		pool.Put(obj)
	}

	fmt.Println("\n当前对象状态:")
	fmt.Printf("已创建对象总数: %d\n", objectCounter)

	fmt.Println("\n=== 模拟第1次GC ===")
	fmt.Println("GC开始...")
	fmt.Println("1. 清空victim（当前为空）")
	fmt.Println("2. 移动active到victim")
	fmt.Println("3. 清空active")

	// 模拟GC：手动触发GC
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // 给GC时间

	fmt.Println("\n=== GC后第1次Get ===")
	// GC后第一次Get，应该从victim中获取
	obj := pool.Get().(*TrackedObject)
	fmt.Printf("获取到对象 #%d (来自victim缓存)\n", obj.ID)
	obj.UsedCount++
	pool.Put(obj)

	fmt.Println("\n=== 放入新对象，然后模拟第2次GC ===")
	// 创建新对象
	newObj := pool.Get().(*TrackedObject) // 如果victim还有，会继续用；否则新建
	newObj.UsedCount++
	pool.Put(newObj)

	fmt.Println("\n模拟第2次GC...")
	fmt.Println("1. 清空victim（包含第1轮的对象）")
	fmt.Println("2. 移动active到victim")
	fmt.Println("3. 清空active")

	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n=== 观察对象存活周期 ===")
	fmt.Println("第1轮的对象已经存活了2个GC周期")
	fmt.Println("如果第2次GC后还需要它们，就会从victim中再次复活")
}

// 验证victim缓存在高频率GC下的效果
func highFrequencyGCDemo() {
	fmt.Println("\n=== 高频GC场景演示 ===")

	pool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}

	// 统计
	var newCreated int
	var mu sync.Mutex

	// 启动多个goroutine频繁使用pool
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				// 获取对象
				obj := pool.Get().([]byte)

				mu.Lock()
				// 判断对象来源（通过容量判断，新创建的是1024）
				if cap(obj) == 1024 {
					// 可能是新创建，也可能是victim中的（victim中的也是1024）
					// 我们通过一个标志位来区分（实际代码中无法区分，这里只演示概念）
					newCreated++
				}
				mu.Unlock()

				// 使用对象
				obj[0] = byte(j)

				// 放回
				pool.Put(obj)

				// 频繁触发GC（模拟场景）
				if j%10 == 0 {
					runtime.GC()
				}

				time.Sleep(time.Microsecond * 100)
			}
		}(i)
	}

	wg.Wait()

	fmt.Printf("总操作次数: %d\n", 10*100)
	fmt.Printf("对象复用情况: 由于victim缓存，很多对象避免了重复创建\n")
}

// 对比有无victim缓存的区别
func compareWithAndWithoutVictim() {
	fmt.Println("\n=== 对比：有victim缓存 vs 无victim缓存 ===")

	// 场景1：有victim缓存（标准sync.Pool）
	fmt.Println("\n场景1: 有victim缓存的sync.Pool")
	pool1 := &sync.Pool{
		New: func() interface{} {
			fmt.Print("N") // N表示New被调用
			return make([]byte, 1024)
		},
	}

	// 模拟GC周期
	fmt.Print("GC周期1: ")
	// 放入对象
	pool1.Put(pool1.Get())
	pool1.Put(pool1.Get())

	// GC
	runtime.GC()
	time.Sleep(50 * time.Millisecond)

	// GC后获取
	fmt.Print("\nGC后获取: ")
	pool1.Get()
	pool1.Get()
	pool1.Get() // 第三个会触发New

	// 再次GC
	fmt.Print("\nGC周期2: ")
	runtime.GC()
	time.Sleep(50 * time.Millisecond)

	fmt.Print("\nGC后获取: ")
	pool1.Get() // 可能从victim获取

	// 场景2：模拟没有victim缓存的情况
	fmt.Println("\n\n场景2: 模拟无victim缓存（每次GC后都重建）")

	newCount := 0
	noVictimGet := func() []byte {
		newCount++
		fmt.Print("N")
		return make([]byte, 1024)
	}

	fmt.Print("\nGC周期1后获取: ")
	noVictimGet()
	noVictimGet()
	noVictimGet()

	fmt.Printf("\n需要创建 %d 次新对象\n", newCount)
}

func main() {
	fmt.Println("=================== sync.Pool的victim缓存机制 ===================")

	demonstrateVictimCache()
	highFrequencyGCDemo()
	compareWithAndWithoutVictim()

	fmt.Println("\n=== 关键理解 ===")
	fmt.Println("1. victim是'幸存者缓存'，让对象多活一个GC周期")
	fmt.Println("2. 减少频繁GC导致的性能抖动")
	fmt.Println("3. 平衡了内存使用和性能")
	fmt.Println("4. 对于短生命周期但频繁使用的对象特别有效")
}
