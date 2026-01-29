package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

/**
goroutine泄漏检测与避免
泄漏：goroutine完成任务之后没有退出 ，一直占用资源，导致内存泄漏。

避免方法：
使用带缓冲的channel或者context来控制goroutine退出
确保在goroutine中处理完任务后正常退出

方法：
使用pprof工具查看goroutine数量
使用runtime.NumGoroutine() 来获取当前goroutine数量

*/

// 一个会导致goroutine泄露的例子
func leak() {
	ch := make(chan int)
	go func() {
		val := <-ch
		fmt.Println("Received:", val)
		// 这里goroutine会一直等待，直到channel被关闭或有数据写入
	}()
	// 这里没有向ch写入数据，也没有关闭ch，导致上面的goroutine一直阻塞，造成泄露
}

// 修复后的版本
func noLeak() {
	ch := make(chan int)
	done := make(chan struct{})
	go func() {
		val := <-ch
		fmt.Println("Received:", val)
		done <- struct{}{}
	}()
	// 模拟一些工作
	time.Sleep(time.Second)
	ch <- 42
	<-done
}

// 使用context控制goroutine退出
func worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Worker exiting")
			return
		default:
			// 执行任务
			fmt.Println("Working...")
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func main() {
	// 演示泄露
	fmt.Println("Goroutine count at start:", runtime.NumGoroutine())
	leak()
	time.Sleep(time.Second) // 等待goroutine创建
	fmt.Println("Goroutine count after leak:", runtime.NumGoroutine())

	// 修复版本
	noLeak()
	fmt.Println("Goroutine count after noLeak:", runtime.NumGoroutine())

	// 使用context
	ctx, cancel := context.WithCancel(context.Background())
	go worker(ctx)
	time.Sleep(2 * time.Second)
	cancel() // 发送取消信号
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Goroutine count after context cancel:", runtime.NumGoroutine())
}
