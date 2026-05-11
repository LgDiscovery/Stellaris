package main

import "fmt"

func generator(done <-chan struct{}, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}()
	return out
}

func main() {
	done := make(chan struct{})
	defer close(done)
	out := generator(done, 2, 3, 4, 5)
	fmt.Println(<-out)
}
