package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

func work(dataChan <-chan int, workerNum int) {
	for n := range dataChan {
		fmt.Printf("worker %d:\t%d\n", workerNum, n)
	}
}

func main() {
	// Шаг 1: узнать количество горутин
	var nFlag = flag.Int("n", 1234, "goroutines amount")
	flag.Parse()

	n := *nFlag

	// Шаг 2: приготовиться словить shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Шаг 3: подготовиться к началу работы: создать канал
	dataChan := make(chan int)

	// Шаг 4: подготовить N worker
	for i := 0; i < n; i++ {
		go work(dataChan, i)
	}

	// Шаг 5: работать, пока не словим shutdown
out:
	for {
		select {
		case <-ctx.Done():
			// словили
			break out
		default:
			break
		}
		// запись данных каждые 500мс
		dataChan <- rand.Intn(10)
		time.Sleep(time.Millisecond * 500)
	}

	// Шаг 6: graceful shutdown (закрыть канал на стороне писателя)
	close(dataChan)

	fmt.Println("done!")
}
