package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
)

// асинхрон по приколу здесь, но он не мешает работать корректно

// sortSlice сортирует слайс, изменяя его
// по задумке функция вызывается рекурсивно и сортирует участки одного и того же массива, на который ссылается слайс
func sortSlice(wg *sync.WaitGroup, a []int) {
	defer wg.Done()

	// рекурсия не применяется для последовательности из 0 или 1 символов
	if len(a) < 2 {
		return
	}

	// выбрать опорный элемент
	base := a[0]

	// это для перемещения назад чисел меньших чем база
	baseIndex := 0
	var tmp int

	for i := 1; i < len(a); i++ {
		if a[i] < base {
			// если встретили число меньше базового переставляем его перед ним (и всей правой частью)
			tmp = a[i]
			copy(a[baseIndex+1:i+1], a[baseIndex:i])
			a[baseIndex] = tmp
			baseIndex += 1
		}
	}

	wg.Add(1)

	// вызываем сорт для частей слева/справа от базы
	go sortSlice(wg, a[:baseIndex])
	if baseIndex < len(a)-1 {
		wg.Add(1)
		go sortSlice(wg, a[baseIndex+1:])
	}
}

func quickSort(a []int) []int {
	result := make([]int, len(a))
	copy(result, a)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	// проводим магию
	sortSlice(wg, result)

	wg.Wait()

	return result
}

func main() {
	nFlag := flag.Int("n", 100, "amount of numbers to sort")
	flag.Parse()

	n := *nFlag

	a := make([]int, n)
	for i := 0; i < n; i++ {
		a[i] = rand.Intn(n*2) - n
	}

	fmt.Println(a)

	sorted := quickSort(a)

	for i := 0; i < n-1; i++ {
		if sorted[i] > sorted[i+1] {
			panic("fix sort!")
		}
	}

	fmt.Println(sorted)
	fmt.Println("success sort!")

	/*
		[41 -30 -52 84 -71 -10 -8 15 -91 -30 18 12 83 96 7 -37 -22 36 92 66 85 -27 52 -60 96 -45 -9 -11 -63 38 -29 -78 -64 67 -29 -23 -62 -19 71 -47 40 39 59 46 80 86 -27 -85 72 -19 -51 75 -24 9 77 -65 -60 -33 58 29 7 52 -16 -5 -85 -42 9 42 86 -44 2 39 71 23 -9 -59 -49 86 -10 60 -82 -95 98 95 -65 -5 27 16 -96 71 -43 64 -57 83 17 29 -90 -9 -40 -55]
		[-96 -95 -91 -90 -85 -85 -82 -78 -71 -65 -65 -64 -63 -62 -60 -60 -59 -57 -55 -52 -51 -49 -47 -45 -44 -43 -42 -40 -37 -33 -30 -30 -29 -29 -27 -27 -24 -23 -22 -19 -19 -16 -11 -10 -10 -9 -9 -9 -8 -5 -5 2 7 7 9 9 12 15 16 17 18 23 27 29 29 36 38 39 39 40 41 42 46 52 52 58 59 60 64 66 67 71 71 71 72 75 77 80 83 83 84 85 86 86 86 92 95 96 96 98]
		success sort!
	*/
}
