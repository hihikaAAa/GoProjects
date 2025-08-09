package main

import (
	"fmt"
	"time"
)

func gather(funcs []func() any) []any {
	type item struct{
		index int
		numb any
	}
	stream := make(chan item, len(funcs))
	done := make(chan struct{})
	answer := make([]any,len(funcs))
	for i, curr := range funcs {
		go func() {
			stream <- item{i,curr()}			
		}()
	}
	go func() {
		for p := range stream {
			answer[p.index] = p.numb
			done <- struct{}{}
		}
		close(stream)
	}()

	for range len(funcs) {
		<-done
	}
	return answer

}


func squared(n int) func() any {
	return func() any {
		time.Sleep(time.Duration(n) * 100 * time.Millisecond)
		return n * n
	}
}

func main() {
	funcs := []func() any{squared(1), squared(2), squared(3), squared(4), squared(5)}

	start := time.Now()
	nums := gather(funcs)
	elapsed := float64(time.Since(start)) / 1_000_000

	fmt.Println(nums)
	fmt.Printf("Took %.0f ms\n", elapsed)
}
