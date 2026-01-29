package main

import (
	"fmt"
	"unsafe"
)

/**
内存对齐是指数据在内存中的地址是其大小的整数倍。这可以加快CPU访问内存的速度。

结构体内存对齐：

结构体的字段按照声明顺序存储，但可能会因为内存对齐而插入填充字节。

结构体的对齐倍数是其所有字段中最大对齐倍数的整数倍。
*/
// 内存对齐示例

// 未优化的结构体
type BadStruct struct {
	a bool  // 1字节
	b int64 // 8字节
	c int32 // 4字节
	d bool  // 1字节
}

// 优化后的结构体
type GoodStruct struct {
	b int64 // 8字节
	c int32 // 4字节
	a bool  // 1字节
	d bool  // 1字节
	// 自动填充2字节
}

// 紧凑排列但可能不对齐
type PackedStruct struct {
	a bool  // 1
	d bool  // 1
	c int32 // 4
	b int64 // 8
}

func memoryAlignmentDemo() {
	bad := BadStruct{}
	good := GoodStruct{}
	packed := PackedStruct{}

	fmt.Println("=== 内存对齐分析 ===")
	fmt.Printf("BadStruct 大小: %d 字节\n", unsafe.Sizeof(bad))
	fmt.Printf("GoodStruct 大小: %d 字节\n", unsafe.Sizeof(good))
	fmt.Printf("PackedStruct 大小: %d 字节\n", unsafe.Sizeof(packed))

	fmt.Println("\n=== 偏移量分析 ===")
	fmt.Println("BadStruct:")
	fmt.Printf("  a offset: %d\n", unsafe.Offsetof(bad.a))
	fmt.Printf("  b offset: %d\n", unsafe.Offsetof(bad.b))
	fmt.Printf("  c offset: %d\n", unsafe.Offsetof(bad.c))
	fmt.Printf("  d offset: %d\n", unsafe.Offsetof(bad.d))

	fmt.Println("\nGoodStruct:")
	fmt.Printf("  b offset: %d\n", unsafe.Offsetof(good.b))
	fmt.Printf("  c offset: %d\n", unsafe.Offsetof(good.c))
	fmt.Printf("  a offset: %d\n", unsafe.Offsetof(good.a))
	fmt.Printf("  d offset: %d\n", unsafe.Offsetof(good.d))
}

// 缓存行优化示例
const cacheLineSize = 64 // 通常为64字节

type FalseSharingBad struct {
	x int64
	y int64
}

type FalseSharingGood struct {
	x int64
	_ [cacheLineSize - 8]byte // 填充到缓存行大小
	y int64
}

func falseSharingDemo() {
	fmt.Println("\n=== 伪共享优化 ===")

	bad := FalseSharingBad{}
	good := FalseSharingGood{}

	fmt.Printf("FalseSharingBad 大小: %d 字节\n", unsafe.Sizeof(bad))
	fmt.Printf("FalseSharingGood 大小: %d 字节\n", unsafe.Sizeof(good))

	// 在多线程环境中，FalseSharingGood可以避免伪共享
}

// 切片内存布局
func sliceMemoryLayout() {
	fmt.Println("\n=== 切片内存布局 ===")

	slice := make([]int, 10, 20)

	// 获取切片底层结构
	sliceHeader := (*[3]uintptr)(unsafe.Pointer(&slice))

	fmt.Printf("切片指针: %p\n", slice)
	fmt.Printf("底层数组指针: %p\n", &slice[0])
	fmt.Printf("长度: %d\n", sliceHeader[0])
	fmt.Printf("容量: %d\n", sliceHeader[1])
	fmt.Printf("数组指针: 0x%x\n", sliceHeader[2])
}

func main() {
	memoryAlignmentDemo()
	falseSharingDemo()
	sliceMemoryLayout()
}
