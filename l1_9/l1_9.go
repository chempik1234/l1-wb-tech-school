package main

import (
	"fmt"
	"sync"
)

// work возводит числа из первого канала в квадрат и кладёт во второй канал
func work(wg *sync.WaitGroup, input <-chan int, output chan<- int) {
	defer wg.Done()
	for n := range input {
		output <- n * n
	}
	close(output)
}

// printFromChannel выводит значения из канала
func printFromChannel(wg *sync.WaitGroup, input <-chan int) {
	defer wg.Done()
	for n := range input {
		fmt.Println("received:", n)
	}
	fmt.Println("finished printing from channel")
}

func main() {
	// дан массив
	n := []int{2, 5, 6, 8, 1, -9, -15}

	fmt.Println("numbers", n)

	// создаём 2 канала
	//
	// main	горутина пишет в nChan		(1 канал) и закрывает его
	// 2	горутина пишет в sqrChan	(2 канал) и закрывает его
	// 3	горутина выводит в stdout

	// каналы закрываются отправителем, чтение корректно завершается

	nChan := make(chan int)
	sqrChan := make(chan int)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// запускаем 2 и 3 горутины
	go work(wg, nChan, sqrChan)
	go printFromChannel(wg, sqrChan)

	// пишем числа в 1 канал
	for _, v := range n {
		nChan <- v
	}
	close(nChan)

	// ждём завершения работы горутин для корректного выключения
	wg.Wait()

	/*
		numbers [2 5 6 8 1 -9 -15]
		received: 4
		received: 25
		received: 36
		received: 64
		received: 1
		received: 81
		received: 225
		finished printing from channel
	*/
}
