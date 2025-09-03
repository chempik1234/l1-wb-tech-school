package main

import (
	"flag"
	"fmt"
	"sync"
	"sync/atomic"
)

type CounterStruct struct {
	counter int32
}

// pointer receiver ибо счётчик у нас один общий

// Increment увеличивает счетчик через atomic
func (s *CounterStruct) Increment() {
	atomic.AddInt32(&s.counter, 1)
}

// Count получает счетчик, тоже через atomic
func (s *CounterStruct) Count() int32 {
	return atomic.LoadInt32(&s.counter)
}

// work увеличивает счётчик у указанной структуры n раз
func work(wg *sync.WaitGroup, s *CounterStruct, n int) {
	defer wg.Done()
	for i := 0; i < n; i++ {
		s.Increment()
	}
}

func main() {
	nFlag := flag.Int("n", 100, "amount of goroutines")
	cFlag := flag.Int("c", 100, "increments per goroutine")
	flag.Parse()

	n := *nFlag
	c := *cFlag

	wg := &sync.WaitGroup{}
	wg.Add(n)

	counterStruct := &CounterStruct{}

	for i := 0; i < n; i++ {
		go work(wg, counterStruct, c)
	}

	wg.Wait()

	fmt.Printf("Workers: %d, Increments per worker: %d, Excepted counter: %d, Actual counter: %d",
		n, c, n*c, counterStruct.Count())

	/*
		go run l1_18.go -n 55 -c 96
		Workers: 55, Increments per worker: 96, Excepted counter: 5280, Actual counter: 5280
	*/
}
