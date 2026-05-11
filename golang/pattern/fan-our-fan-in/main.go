package main

import (
	"fmt"
	"sync"
)

func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func fanOut(in <-chan int, workers int) []<-chan int {
	channels := make([]<-chan int, workers)
	for i := 0; i < workers; i++ {
		channels[i] = square(in)
	}
	return channels
}

func fanIn(inputs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	merged := make(chan int)

	output := func(c <-chan int) {
		defer wg.Done()
		for n := range c {
			merged <- n
		}
	}
	wg.Add(len(inputs))
	for _, input := range inputs {
		go output(input)
	}
	go func() {
		wg.Wait()
		close(merged)
	}()
	return merged
}

func main() {
	in := generate(2, 3, 4, 5, 6, 7, 8, 9)
	workers := fanOut(in, 3)
	out := fanIn(workers...)
	for n := range out {
		fmt.Println(n)
	}
}
