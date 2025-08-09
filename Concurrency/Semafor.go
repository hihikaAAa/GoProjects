package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func say(id int, phrase string) {
	for _, word := range strings.Fields(phrase) {
		fmt.Printf("Worker #%d says: %s...\n", id, word)
		dur := time.Duration(rand.Intn(100)) * time.Millisecond
		time.Sleep(dur)
	}
}

func makePool(n int, handler func(int, string)) (func(string), func()) {
	pool := make(chan int, n)
	for i := range n{
		pool <- i+1
	}
	handle := func(i string) {
		token := <-pool
		go func(){
			handler(token, i)
			pool <- token
		}()
	}

	wait := func() {
		for range n{
			<-pool
		}
		close(pool)
		
	}


	return handle, wait
}



func main() {
	phrases := []string{
		"go is awesome",
		"cats are cute",
		"rain is wet",
		"channels are hard",
		"floor is lava",
	}
	n := 2
	handle, wait := makePool(n, say)
	for _, phrase := range phrases {
		handle(phrase)
	}
	wait()
}

