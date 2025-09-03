package main

import (
	"fmt"
	"slices"
)

// intersection находит пересечение с сохранением уникальности
func intersection(a []int, b []int) []int {
	result := make([]int, 0)

	// будем проходить по множеству наименьшей длины, и сверять с числами из "другого", "второго" множества
	var iteratedSet *[]int
	var secondSet *[]int
	if len(a) < len(b) {
		iteratedSet = &a
		secondSet = &b
	} else {
		iteratedSet = &b
		secondSet = &a
	}

	// result меньше или равен по длине наименьшему множеству,
	// поэтому на нём использовать Contains лучше,
	// чем на secondSet, который >= ему по длине

	for _, v := range *iteratedSet {
		// если число уже найдено то его не надо писать ещё раз
		// выйдем сразу - не будем по 100 раз проверять "есть ли число 2 в другом множестве"
		if slices.Contains(result, v) {
			continue
		}

		// поиск чисел через банальнейший contains
		//
		// добавление в результирующий слайсик,
		//   если числа ещё в нем нет
		//   и если оно встречается в "другом" "множестве"
		if slices.Contains(*secondSet, v) {
			result = append(result, v)
		}
	}
	return result
}

func main() {
	// можно взять любые числа
	a := []int{1, 2, 3, 10}
	b := []int{2, 3, 4, 5, 10}

	fmt.Println(a, b, "пересечение", intersection(a, b))

	// [1 2 3 10] [2 3 4 5 10] пересечение [2 3 10]
}
