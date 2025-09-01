package main

/*
#include <string.h>
#include <stdlib.h>
#include <windows.h>
#include <stdint.h>

void corrupt_memory(void* ptr, size_t size) {
    memset(ptr, 0xFF, size); // Заполняем мусором память
}

uintptr_t get_goroutine_pointer_win() {
    // В Windows Go использует TEB для хранения указателя на g
    #ifdef _WIN64
        return (uintptr_t)__readgsqword(0x30); // TEB для x64
    #else
        return (uintptr_t)__readfsdword(0x18); // TEB для x86
    #endif
}

uintptr_t get_thread_id() {
    return (uintptr_t)GetCurrentThreadId();
}
*/
import "C"
import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// randError возвращает ошибку с 20% шансом
//
// используется в функциях с названияи формата exit<N><Way>
func randError() error {
	if rand.Intn(5) == 1 {
		return errors.New("random number == 1: time to exit")
	}
	return nil
}

// 1 способ по условию: рандомно возникающая ошибка
//
// если бы функция выходила при чтении из канала, создавался бы Deadlock, а так внутреннее условие никому не мешает
func exit1WithCondition(wg *sync.WaitGroup) {
	defer wg.Done()
	i := 0
	for {
		// условие выход при первой ошибке
		err := randError()
		if err != nil {
			fmt.Printf("func 1:\terror\t\t%d: %v\n", i, err)
			break
		}
		i++

		fmt.Printf("func 1:\tcontinue\t%d\n", i)
	}
	fmt.Println("func 1:\tend!")
}

// 2 способ канал с уведомлениями: принимает канал при чтении из которого горутина завершается, по завершении она также
// отправляет значение в созданный канал с уведомлениями о завершениями
//
// от этой обёртки вокруг exit2Goroutine в теории можно избавиться и создавать канал в main
func exit2WithChannel(wg *sync.WaitGroup, inputChan <-chan int, signal <-chan struct{}) chan struct{} {
	sendOnFinished := make(chan struct{})
	go exit2Goroutine(wg, inputChan, signal, sendOnFinished)
	return sendOnFinished
}

func exit2Goroutine(wg *sync.WaitGroup, inputChan <-chan int, signal <-chan struct{}, sendOnFinished chan<- struct{}) {
	defer wg.Done()
out:
	for {
		select {
		case n := <-inputChan:
			fmt.Printf("func 2:\tinputChan\t%d\n", n)
		case s := <-signal:
			fmt.Printf("func 2:\tsignal\t%v\n", s)
			break out
		}
	}
	sendOnFinished <- struct{}{}
	fmt.Println("func 2:\tend!")
}

// 3 способ контекст: в данном примере context, однако ctx.Done() это ведь тоже канал с уведомлениями
func exit3WithContext(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("func 3:\tok, waiting for end")
	<-ctx.Done()
	fmt.Println("func 3:\tend!")
}

// 4 способ: функция runtime.Goexit() завершает горутину
func exit4WithRuntimeGoexit(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("func 4:\tcalling runtime.Goexit() (if runtime.Goexit() works you won't see anything else)")
	runtime.Goexit()
	fmt.Println("func 4:\thow did you see that!!!!")
}

func getCurrentGoroutinePointer() unsafe.Pointer {
	// fmt.Println(unsafe.Pointer(uintptr(C.get_goroutine_pointer_win())))
	// fmt.Println(unsafe.Pointer(uintptr(C.get_thread_id())))
	return unsafe.Pointer(uintptr(C.get_thread_id())) // getg()
}

func exit5WithDestroyingGoroutineByPointer() {
	// текущая горутина
	ptr := getCurrentGoroutinePointer()
	fmt.Printf("func 5:\tpointer to goroutine: %v\n", ptr)

	// Затираем область памяти (чисто теоретически чучут)

	fmt.Println("func 5:\tfire CGO!")

	C.corrupt_memory(ptr, C.size_t(40))

	fmt.Println("func 5:\thow was that?")
}

func main() {
	fmt.Println("-- Begin 1: exit on condition")
	wg1 := &sync.WaitGroup{}
	wg1.Add(1)
	go exit1WithCondition(wg1)
	wg1.Wait()
	fmt.Println("-- Finish 1")

	runtime.GC()

	fmt.Println("-- Begin 2: exit on channel")
	wg2 := &sync.WaitGroup{}
	wg2.Add(1)
	input2 := make(chan int)
	signal2 := make(chan struct{})
	finished2 := exit2WithChannel(wg2, input2, signal2)

loop2:
	for i := 0; i < 10; i++ {
		/*
			if i <= 5 {
				input2 <- i
				if i == 5 {
					signal2 <- struct{}{}
				}
			} else {
				<-finished2
				fmt.Println("-- Finish 2: returned")
				break loop2
			}*/
		select {
		case <-finished2:
			fmt.Println("-- Finish 2: returned")
			break loop2
		case input2 <- i:
			if i == 5 {
				signal2 <- struct{}{}
			}
		}
	}
	close(input2)
	close(signal2)

	wg2.Wait()

	runtime.GC()

	fmt.Println("-- Begin 3: exit on context")
	wg3 := &sync.WaitGroup{}
	wg3.Add(1)
	ctx3, cancel3 := context.WithTimeout(context.Background(), time.Millisecond*500)
	go exit3WithContext(ctx3, wg3)
	<-ctx3.Done()
	cancel3()
	wg3.Wait()

	runtime.GC()

	fmt.Println("-- Begin 4: exit on runtime.Goexit()")
	wg4 := &sync.WaitGroup{}
	wg4.Add(1)
	go exit4WithRuntimeGoexit(wg4)
	wg4.Wait()
	fmt.Println("-- Finish 4")

	runtime.GC()

	fmt.Println("-- Begin 5: CGO corruption")
	go exit5WithDestroyingGoroutineByPointer()
	time.Sleep(2 * time.Second)

	fmt.Println("-- Finish L1.6!")
}
