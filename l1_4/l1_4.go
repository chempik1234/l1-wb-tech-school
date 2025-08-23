package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"
)

// 1 способ остановки горутин: предоставить общий контекст (который отменится при SIGINT)
func workWithContext(wg *sync.WaitGroup, ctx context.Context, workerNum int) {
	if wg != nil {
		defer wg.Done()
	}

	fmt.Printf("ctx worker %d:\tstart\n", workerNum)
	<-ctx.Done()
	fmt.Printf("ctx worker %d:\tdone\n", workerNum)
}

// 2 способ остановки горутин: дать им канал для работы и закрыть его,
// чтобы заставить завершиться
func work(wg *sync.WaitGroup, dataChan <-chan int, workerNum int) {
	if wg != nil {
		defer wg.Done()
	}

	for n := range dataChan {
		fmt.Printf("worker %d:\t%d\n", workerNum, n)
	}
	fmt.Printf("worker %d:\tdone\n", workerNum)
}

// 3 способ остановки: дать канал специально для завершения (как ctx.Done())
func work3(wg *sync.WaitGroup, dataChan <-chan int, stopChan <-chan struct{}, workerNum int) {
	if wg != nil {
		defer wg.Done()
	}

	fmt.Printf("work3 worker %d:\tstart\n", workerNum)
	var num int

work3:
	for {
		// Сделаем 2 select для того чтобы stopChan был в приоритете:
		// один select не гарантирует порядок чтения
		select {
		case <-stopChan:
			break work3
		default:
			break
		}
		// ... если пока завершение не требуется, продолжаем обрабатывать
		// однако чтение stopChan убирать не следует: пока мы "вечность" ждём dataChan,
		// сообщение из stopChan, возможно, уже придёт
		select {
		case <-stopChan:
			break work3
		case num = <-dataChan:
			fmt.Printf("work3 worker %d:\t%d\n", workerNum, num)
		}

	}
	fmt.Printf("work3 worker %d:\tdone\n", workerNum)
}

func main() {
	var nFlag = flag.Int("n", 3, "goroutines amount")
	flag.Parse()

	n := *nFlag

	// способ обработки Ctrl + C - NotifyContext с сигналом os.Interrupt
	// при SIGINT будет вызван cancel()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	dataChan := make(chan int)

	stopChan := make(chan struct{})

	// Чтобы успешно завершить все горутины и только потом выйти, создадим WG
	wg := &sync.WaitGroup{}
	wg.Add(n * 3)

	for i := 0; i < n; i++ {
		go work(wg, dataChan, i)
	}
	for i := 0; i < n; i++ {
		go workWithContext(wg, ctx, i)
	}
	for i := 0; i < n; i++ {
		go work3(wg, dataChan, stopChan, i)
	}

	// работать, пока не словим shutdown
out:
	for {
		select {
		case <-ctx.Done():
			// словили
			break out
		default:
			break
		}
		dataChan <- rand.Intn(10)
		time.Sleep(time.Millisecond * 500)
	}

	close(stopChan)
	// Если бы мы просто отправили значение в канал:
	//
	// stopChan <- struct{}{}
	//
	// Оно бы было прочтено всего единожды, а так не пойдёт

	// graceful shutdown (закрыть канал на стороне писателя: см. func work)
	close(dataChan)

	// нельзя сказать в каком порядке завершатся горутины, но повод для этого
	// появляется сначала для workWithContext (<-ctx.Done())
	// и лишь затем для work (close(dataChan))

	wg.Wait()

	fmt.Println("done!")
}
