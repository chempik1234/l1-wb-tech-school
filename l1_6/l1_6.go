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

/*
Пример вывода

-- Begin 1: exit on condition
func 1: continue        1
func 1: continue        2
func 1: continue        3
func 1: continue        4
func 1: continue        5
func 1: continue        6
func 1: continue        7
func 1: continue        8
func 1: continue        9
func 1: error           9: random number == 1: time to exit
func 1: end!
-- Finish 1
-- Begin 2: exit on channel
func 2: inputChan       0
func 2: inputChan       1
func 2: inputChan       2
func 2: inputChan       3
func 2: inputChan       4
func 2: inputChan       5
func 2: signal  {}
func 2: end!
-- Finish 2: returned
-- Begin 3: exit on context
func 3: ok, waiting for end
func 3: end!
-- Begin 4: exit on runtime.Goexit()
func 4: calling runtime.Goexit() (if runtime.Goexit() works you won't see anything else)
-- Finish 4
-- Begin 5: CGO corruption
func 5: pointer to goroutine: 0x5990
func 5: fire CGO!
Exception 0xc0000005 0x1 0x5998 0x7ffbd871d794
PC=0x7ffbd871d794
signal arrived during external code execution

runtime.cgocall(0x7ff76608f640, 0xc0000bbf60)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/cgocall.go:167 +0x3e fp=0xc0000bbf38 sp=0xc0000bbed0 pc=0x7ff7660492fe
main._Cfunc_corrupt_memory(0x5990, 0x28)
        _cgo_gotypes.go:52 +0x48 fp=0xc0000bbf60 sp=0xc0000bbf38 pc=0x7ff76608e368
main.exit5WithDestroyingGoroutineByPointer.func1(...)
        C:/Users/Danis/wbtech/l1/l1_6/l1_6.go:126
main.exit5WithDestroyingGoroutineByPointer()
        C:/Users/Danis/wbtech/l1/l1_6/l1_6.go:126 +0xd6 fp=0xc0000bbfe0 sp=0xc0000bbf60 pc=0x7ff76608ee96
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc0000bbfe8 sp=0xc0000bbfe0 pc=0x7ff766051d81
created by main.main in goroutine 1
        C:/Users/Danis/wbtech/l1/l1_6/l1_6.go:199 +0x536

goroutine 1 gp=0xc0000021c0 m=nil [sleep]:
runtime.gopark(0x145790f42544?, 0x7ff76604fd97?, 0xe8?, 0xdd?, 0x7ff76602671f?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc00005dda0 sp=0xc00005dd80 pc=0x7ff76604b22e
time.Sleep(0x77359400)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/time.go:338 +0x167 fp=0xc00005ddf8 sp=0xc00005dda0 pc=0x7ff76604e107
main.main()
        C:/Users/Danis/wbtech/l1/l1_6/l1_6.go:200 +0x545 fp=0xc00005df50 sp=0xc00005ddf8 pc=0x7ff76608f445
runtime.main()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:283 +0x27d fp=0xc00005dfe0 sp=0xc00005df50 pc=0x7ff76601c9fd
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc00005dfe8 sp=0xc00005dfe0 pc=0x7ff766051d81

goroutine 2 gp=0xc0000028c0 m=nil [force gc (idle)]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc000059fa8 sp=0xc000059f88 pc=0x7ff76604b22e
runtime.goparkunlock(...)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:441
runtime.forcegchelper()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:348 +0xb8 fp=0xc000059fe0 sp=0xc000059fa8 pc=0x7ff76601cd18
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc000059fe8 sp=0xc000059fe0 pc=0x7ff766051d81
created by runtime.init.7 in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:336 +0x1a

goroutine 3 gp=0xc000002c40 m=nil [GC sweep wait]:
runtime.gopark(0x1?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc00005bf80 sp=0xc00005bf60 pc=0x7ff76604b22e
runtime.goparkunlock(...)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:441
runtime.bgsweep(0xc000022380)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgcsweep.go:316 +0xdf fp=0xc00005bfc8 sp=0xc00005bf80 pc=0x7ff766006c7f
runtime.gcenable.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:204 +0x25 fp=0xc00005bfe0 sp=0xc00005bfc8 pc=0x7ff765ffb0c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc00005bfe8 sp=0xc00005bfe0 pc=0x7ff766051d81
created by runtime.gcenable in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:204 +0x66

goroutine 4 gp=0xc000002e00 m=nil [sleep]:
runtime.gopark(0xc00003a0c0?, 0x1457cc96d844?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc00006df38 sp=0xc00006df18 pc=0x7ff76604b22e
runtime.goparkunlock(...)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:441
runtime.(*scavengerState).sleep(0x7ff76618b7c0, 0x4113880000000000)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgcscavenge.go:504 +0xfb fp=0xc00006dfa8 sp=0xc00006df38 pc=0x7ff76600489b
runtime.bgscavenge(0xc000022380)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgcscavenge.go:662 +0x74 fp=0xc00006dfc8 sp=0xc00006dfa8 pc=0x7ff766004c94
runtime.gcenable.gowrap2()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:205 +0x25 fp=0xc00006dfe0 sp=0xc00006dfc8 pc=0x7ff765ffb065
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc00006dfe8 sp=0xc00006dfe0 pc=0x7ff766051d81
created by runtime.gcenable in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:205 +0xa5

goroutine 5 gp=0xc000003340 m=nil [finalizer wait]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc00006fe30 sp=0xc00006fe10 pc=0x7ff76604b22e
runtime.runfinq()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mfinal.go:196 +0x107 fp=0xc00006ffe0 sp=0xc00006fe30 pc=0x7ff765ffa0c7
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc00006ffe8 sp=0xc00006ffe0 pc=0x7ff766051d81
created by runtime.createfing in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mfinal.go:166 +0x3d

goroutine 18 gp=0xc000003500 m=nil [GC worker (idle)]:
runtime.gopark(0x2?, 0x0?, 0x28?, 0x43?, 0x9?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc000069f38 sp=0xc000069f18 pc=0x7ff76604b22e
runtime.gcBgMarkWorker(0xc000096070)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1423 +0xe9 fp=0xc000069fc8 sp=0xc000069f38 pc=0x7ff765ffd5e9
runtime.gcBgMarkStartWorkers.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x25 fp=0xc000069fe0 sp=0xc000069fc8 pc=0x7ff765ffd4c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc000069fe8 sp=0xc000069fe0 pc=0x7ff766051d81
created by runtime.gcBgMarkStartWorkers in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x105

goroutine 19 gp=0xc0000861c0 m=nil [GC worker (idle)]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc00009ff38 sp=0xc00009ff18 pc=0x7ff76604b22e
runtime.gcBgMarkWorker(0xc000096070)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1423 +0xe9 fp=0xc00009ffc8 sp=0xc00009ff38 pc=0x7ff765ffd5e9
runtime.gcBgMarkStartWorkers.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x25 fp=0xc00009ffe0 sp=0xc00009ffc8 pc=0x7ff765ffd4c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc00009ffe8 sp=0xc00009ffe0 pc=0x7ff766051d81
created by runtime.gcBgMarkStartWorkers in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x105

goroutine 20 gp=0xc000086380 m=nil [GC worker (idle)]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc0000a1f38 sp=0xc0000a1f18 pc=0x7ff76604b22e
runtime.gcBgMarkWorker(0xc000096070)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1423 +0xe9 fp=0xc0000a1fc8 sp=0xc0000a1f38 pc=0x7ff765ffd5e9
runtime.gcBgMarkStartWorkers.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x25 fp=0xc0000a1fe0 sp=0xc0000a1fc8 pc=0x7ff765ffd4c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc0000a1fe8 sp=0xc0000a1fe0 pc=0x7ff766051d81
created by runtime.gcBgMarkStartWorkers in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x105

goroutine 21 gp=0xc000086540 m=nil [GC worker (idle)]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc00009bf38 sp=0xc00009bf18 pc=0x7ff76604b22e
runtime.gcBgMarkWorker(0xc000096070)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1423 +0xe9 fp=0xc00009bfc8 sp=0xc00009bf38 pc=0x7ff765ffd5e9
runtime.gcBgMarkStartWorkers.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x25 fp=0xc00009bfe0 sp=0xc00009bfc8 pc=0x7ff765ffd4c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc00009bfe8 sp=0xc00009bfe0 pc=0x7ff766051d81
created by runtime.gcBgMarkStartWorkers in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x105

goroutine 22 gp=0xc000086700 m=nil [GC worker (idle)]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc00009df38 sp=0xc00009df18 pc=0x7ff76604b22e
runtime.gcBgMarkWorker(0xc000096070)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1423 +0xe9 fp=0xc00009dfc8 sp=0xc00009df38 pc=0x7ff765ffd5e9
runtime.gcBgMarkStartWorkers.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x25 fp=0xc00009dfe0 sp=0xc00009dfc8 pc=0x7ff765ffd4c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc00009dfe8 sp=0xc00009dfe0 pc=0x7ff766051d81
created by runtime.gcBgMarkStartWorkers in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x105

goroutine 23 gp=0xc0000868c0 m=nil [GC worker (idle)]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc0000a9f38 sp=0xc0000a9f18 pc=0x7ff76604b22e
runtime.gcBgMarkWorker(0xc000096070)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1423 +0xe9 fp=0xc0000a9fc8 sp=0xc0000a9f38 pc=0x7ff765ffd5e9
runtime.gcBgMarkStartWorkers.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x25 fp=0xc0000a9fe0 sp=0xc0000a9fc8 pc=0x7ff765ffd4c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc0000a9fe8 sp=0xc0000a9fe0 pc=0x7ff766051d81
created by runtime.gcBgMarkStartWorkers in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x105

goroutine 24 gp=0xc000086a80 m=nil [GC worker (idle)]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc0000abf38 sp=0xc0000abf18 pc=0x7ff76604b22e
runtime.gcBgMarkWorker(0xc000096070)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1423 +0xe9 fp=0xc0000abfc8 sp=0xc0000abf38 pc=0x7ff765ffd5e9
runtime.gcBgMarkStartWorkers.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x25 fp=0xc0000abfe0 sp=0xc0000abfc8 pc=0x7ff765ffd4c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc0000abfe8 sp=0xc0000abfe0 pc=0x7ff766051d81
created by runtime.gcBgMarkStartWorkers in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x105

goroutine 25 gp=0xc000086c40 m=nil [GC worker (idle)]:
runtime.gopark(0x145790f42544?, 0x3?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc0000a5f38 sp=0xc0000a5f18 pc=0x7ff76604b22e
runtime.gcBgMarkWorker(0xc000096070)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1423 +0xe9 fp=0xc0000a5fc8 sp=0xc0000a5f38 pc=0x7ff765ffd5e9
runtime.gcBgMarkStartWorkers.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x25 fp=0xc0000a5fe0 sp=0xc0000a5fc8 pc=0x7ff765ffd4c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc0000a5fe8 sp=0xc0000a5fe0 pc=0x7ff766051d81
created by runtime.gcBgMarkStartWorkers in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x105

goroutine 26 gp=0xc000086e00 m=nil [GC worker (idle)]:
runtime.gopark(0x145790f42544?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc0000a7f38 sp=0xc0000a7f18 pc=0x7ff76604b22e
runtime.gcBgMarkWorker(0xc000096070)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1423 +0xe9 fp=0xc0000a7fc8 sp=0xc0000a7f38 pc=0x7ff765ffd5e9
runtime.gcBgMarkStartWorkers.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x25 fp=0xc0000a7fe0 sp=0xc0000a7fc8 pc=0x7ff765ffd4c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc0000a7fe8 sp=0xc0000a7fe0 pc=0x7ff766051d81
created by runtime.gcBgMarkStartWorkers in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x105

goroutine 27 gp=0xc000086fc0 m=nil [GC worker (idle)]:
runtime.gopark(0x145790f42544?, 0x3?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc0000b1f38 sp=0xc0000b1f18 pc=0x7ff76604b22e
runtime.gcBgMarkWorker(0xc000096070)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1423 +0xe9 fp=0xc0000b1fc8 sp=0xc0000b1f38 pc=0x7ff765ffd5e9
runtime.gcBgMarkStartWorkers.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x25 fp=0xc0000b1fe0 sp=0xc0000b1fc8 pc=0x7ff765ffd4c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc0000b1fe8 sp=0xc0000b1fe0 pc=0x7ff766051d81
created by runtime.gcBgMarkStartWorkers in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x105

goroutine 28 gp=0xc000087180 m=nil [GC worker (idle)]:
runtime.gopark(0x145790f42544?, 0x3?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc0000b3f38 sp=0xc0000b3f18 pc=0x7ff76604b22e
runtime.gcBgMarkWorker(0xc000096070)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1423 +0xe9 fp=0xc0000b3fc8 sp=0xc0000b3f38 pc=0x7ff765ffd5e9
runtime.gcBgMarkStartWorkers.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x25 fp=0xc0000b3fe0 sp=0xc0000b3fc8 pc=0x7ff765ffd4c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc0000b3fe8 sp=0xc0000b3fe0 pc=0x7ff766051d81
created by runtime.gcBgMarkStartWorkers in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x105

goroutine 29 gp=0xc000087340 m=nil [GC worker (idle)]:
runtime.gopark(0x145790f42544?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/proc.go:435 +0xce fp=0xc0000adf38 sp=0xc0000adf18 pc=0x7ff76604b22e
runtime.gcBgMarkWorker(0xc000096070)
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1423 +0xe9 fp=0xc0000adfc8 sp=0xc0000adf38 pc=0x7ff765ffd5e9
runtime.gcBgMarkStartWorkers.gowrap1()
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x25 fp=0xc0000adfe0 sp=0xc0000adfc8 pc=0x7ff765ffd4c5
runtime.goexit({})
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc0000adfe8 sp=0xc0000adfe0 pc=0x7ff766051d81
created by runtime.gcBgMarkStartWorkers in goroutine 1
        C:/Users/Danis/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.windows-amd64/src/runtime/mgc.go:1339 +0x105
rax     0x5990
rbx     0xc0000bbf60
rcx     0x59a0
rdx     0x10f
rdi     0x140
rsi     0x7ff76618b900
rbp     0x85d25ffb50
rsp     0x85d25ffb28
r8      0x18
r9      0x20
r10     0x7ffbd8630000
r11     0x7ffbd871d794
r12     0x0
r13     0x24
r14     0xc000186540
r15     0x25
rip     0x7ffbd871d794
rflags  0x10202
cs      0x33
fs      0x53
gs      0x2b
exit status 2
*/
