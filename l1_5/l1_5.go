package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Шаг 1. Определить время в читаемом формате
	var nFlag = flag.Int("n", 2, "execution seconds")
	flag.Parse()

	N := *nFlag

	// Шаг 2. Перевести его в time.Duration
	waitTime := time.Duration(N) * time.Second

	// Шаг 3. Создать рабочий ticker для записи в канал
	tickTime := time.Duration(200) * time.Millisecond
	ticker := time.NewTicker(tickTime)

	// Шаг 4. Работать и ждать
	inputChan := make(chan int)

	go func(ctx context.Context, input chan int) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Done")
				return
			case n := <-inputChan:
				fmt.Printf("Received: %d\n", n)
			}
		}
	}(ctx, inputChan)

	nowTime := time.Now()
	fmt.Println("-- Begin")

	// Есть 2 варианта:
	// 1) ticker.C
	// 2) time.Sleep
	// Sleep может заставить программу работать дольше положенного, например на 1.99 сек вызовется Sleep(0.1s)
	//   и завершение произойдёт на 2.09 сек
	// В цикле использовать time.After вместе с default не получится,
	//   если создавать таймер (что делает time.After под капотом) каждый раз: счётчик будет считаться заново!

	// Шаг 4.1: создадим таймер для отчёта времени завершения до начала выполнения
	afterTimeout := time.After(waitTime)

out:
	for {
		select {
		// Шаг 5. Выйти при достижении времени: time.After помогает определить, достигнуто ли время выполнения
		case <-afterTimeout:
			fmt.Println("-- Time is up!")
			break out
		case <-ticker.C:
			fmt.Printf("waiting, %vs\t", math.Round(time.Since(nowTime).Seconds()*100)/100)
			inputChan <- rand.Intn(10)
		}
	}

	// Шаг 5. завершить воркер
	cancel()

	fmt.Println("-- End")
}
