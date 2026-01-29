package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
)

/**
使用pprof进行CPU和内存分析
pprof是Go语言自带的性能分析工具。

CPU分析：

使用pprof.StartCPUProfile和pprof.StopCPUProfile来记录CPU使用情况。

使用go tool pprof工具查看分析结果。

内存分析：

使用pprof.WriteHeapProfile来记录内存使用情况。

运行程序，然后访问http://localhost:8080/debug/pprof/，可以看到pprof的页面。

我们也可以使用go tool pprof命令行工具：

bash
# CPU分析
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# 内存分析
go tool pprof http://localhost:8080/debug/pprof/heap
*/

func heavyTask() {
	for i := 0; i < 1000000; i++ {
		_ = make([]byte, 1024)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	heavyTask()
	fmt.Fprintf(w, "Hello, world!")
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
