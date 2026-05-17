package main

import (
	"fmt"
	"sync"
)

func main() {
	var count int
	var mutex sync.Mutex
	wg := sync.WaitGroup{}

	for range 5 {
		wg.Go(func() {
			for range 1000 {
				mutex.Lock()
				count++
				mutex.Unlock()
			}
		})
	}

	wg.Wait()

	fmt.Printf("Count: %d\n", count)
}
