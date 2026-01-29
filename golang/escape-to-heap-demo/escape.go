package main

import (
	"fmt"
	"runtime"
)

/**
2.1 逃逸分析
逃逸分析是编译器用来确定变量分配在栈上还是堆上的过程。如果变量在函数外部被引用，则逃逸到堆上，否则分配在栈上。

逃逸场景：

函数返回局部变量的指针。

将局部变量赋值给包级别的变量。

在闭包中引用局部变量。

将局部变量发送到channel或存储在切片中。
*/

// 逃逸分析示例

// 情况1：返回指针，逃逸到堆
func escapeToHeap() *int {
	v := 42 // 逃逸到堆
	return &v
}

// 情况2：大对象，逃逸到堆
func largeObject() {
	_ = make([]int, 10000) // 逃逸到堆
}

// 情况3：闭包引用，逃逸到堆
func closureEscape() func() int {
	x := 100 // 逃逸到堆
	return func() int {
		return x
	}
}

// 情况4：接口类型，逃逸到堆
func interfaceEscape() {
	var x interface{}
	x = 42 // 42逃逸到堆
	_ = x
}

// 情况5：发送到channel，逃逸到堆
func channelEscape() {
	ch := make(chan *int, 1)
	x := 42 // 逃逸到堆
	ch <- &x
}

// 没有逃逸的情况
func noEscape() int {
	v := 42 // 栈上分配
	return v
}

func printMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
	fmt.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	fmt.Printf("\tSys = %v MiB", m.Sys/1024/1024)
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func main() {
	fmt.Println("=== 逃逸分析演示 ===")

	// 编译时查看逃逸分析：go build -gcflags="-m"

	printMemStats()

	// 测试各种逃逸情况
	_ = escapeToHeap()
	largeObject()
	f := closureEscape()
	f()
	interfaceEscape()
	channelEscape()
	noEscape()

	runtime.GC()
	printMemStats()
}
