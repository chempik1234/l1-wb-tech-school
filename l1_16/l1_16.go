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
}
