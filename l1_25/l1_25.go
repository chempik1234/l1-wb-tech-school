package main

import (
	"fmt"
	"time"
)

// здесь используется цикл в котором бесконечно проверяется разница начального и текущего времени
func sleep1(duration time.Duration) {
	start := time.Now()
	for time.Now().Sub(start) < duration {
	}
}

// time.After создаёт таймер и чтение произойдёт только после прошествия времени
func sleep2(duration time.Duration) {
	chanAfter := time.After(duration)
	<-chanAfter
}

func main() {
	duration := time.Second * 5

	fmt.Println("sleep for 5s with cycle and time checking")
	sleep1(duration)
	fmt.Printf("5s passed\n\n")

	fmt.Println("sleep for 5s with timer")
	sleep2(duration)
	fmt.Printf("5s passed\n\n")
}
