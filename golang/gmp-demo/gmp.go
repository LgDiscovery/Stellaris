package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

/**
GMP: 原理
G: 用户态协程，是Go语言中的轻量级线程
M: 内核态线程，由操作系统调度
P: 介于G和M之间，负责调度G在M上运行

M 与 P 绑定 会从 p 维护的队列中获取 G 运行，如果 当前队列没有任务 从全局队列获取 如果全局也为空 则从 其他 p
维护的队列获取 G 运行。 如果 G执行阻塞了，M 也会被阻塞，此时 P 和 M 会解绑 ，并寻找一个空闲的 M (或者是创建新的M)
来继续执行这个P上的G。当阻塞操作完成后，G会被放入全局队列，M会进入休眠状态。

*/

//展示GMP调度的例子

func gmpDemo() {
	//查看当前GOMaxProcs(p数量)
	fmt.Println("GOMAXPROCS=", runtime.GOMAXPROCS(0))
	// 查看当前goroutine数量
	fmt.Println("NumGoroutine=", runtime.NumGoroutine())

	var wg sync.WaitGroup

	//创建大量的goroutine,观察调度
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			//模拟计算任务
			sum := 0
			for j := 0; j < 1000000; j++ {
				sum += j
			}
			// 主动让出CPU
			runtime.Gosched()
			fmt.Printf("Goroutine %d completed\n", id)
		}(i)
	}

	wg.Wait()
}

// 展示p的本地队列和全局队列
func queueDemo() {
	//设置只使用1个p,更容易观察
	runtime.GOMAXPROCS(1)

	done := make(chan bool)

	//创建阻塞的goroutine,占用p
	go func() {
		time.Sleep(2 * time.Second)
		done <- true
	}()

	// 创建多个goroutine 他们会进入队列
	for i := 0; i < 5; i++ {
		go func(id int) {
			fmt.Printf("Goroutine %d started\n", id)
			time.Sleep(1 * time.Second)
			fmt.Printf("Goroutine %d finished\n", id)
		}(i)
	}

	<-done
}

func main() {
	fmt.Println("===GMP调度演示====")
	gmpDemo()
	fmt.Println("\n===队列调度演示====")
	queueDemo()
}
