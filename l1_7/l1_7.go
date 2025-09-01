package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
)

// work принимает общие для вызывающей стороны sync.WaitGroup, sync.Mutex и мапу map[int]int
// work пишет в map 5 ключей от 0 до 4
// по задумке это должно запускаться конкуррентно
func work(wg *sync.WaitGroup, mu *sync.Mutex, a map[int]int) {
	defer wg.Done()
	var value int
	for i := 0; i < 5; i++ {
		value = rand.Intn(100)

		// Заблокировать атомарную часть
		// Вывод входит в неё, так как по нему
		//   мы проверяем хронологию операций!
		mu.Lock()
		a[i] = value
		fmt.Println("--", strings.Repeat("\t", i), i, value)
		mu.Unlock()
	}
}

func main() {
	n := 3
	a := make(map[int]int)

	// обычный Mutex подойдёт для постоянной записи
	mu := &sync.Mutex{}

	// запустим n воркеров и посмотрим на вывод:
	// в каждой колонке нужно взять самое нижнее число
	// смотрите в конце
	wg := &sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go work(wg, mu, a)
	}

	// дождёмся окончания записи
	wg.Wait()

	// сравним результат с тем, что видели:
	fmt.Println(a)

	/*
		--  0 48
		--       1 35
		--               2 31
		--                       3 17
		--                               4 95
		--  0 70
		--       1 91
		--               2 94
		--                       3 93
		--                               4 49
		--  0 24
		--       1 29
		--               2 29
		--                       3 90
		--                               4 48
		map[0:24 1:29 2:29 3:90 4:48]

		24=24, 29=29, 29=29, 90=90, 48=48, всё круто!

		С мьютексом гонок нет

		Если бы не было мьютекса, то числа бы не совпадали, например:

		map[0:48, 1:91, 2:29, 3:90, 4:49]
		Found 2 data race(s)
		exit status 66
	*/
}
