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

// intersectionWithMap делает то же самое что ли intersection,
// но сохраняет значения как ключи в мапе
// их уникальность легко считается
//
// все значения пишутся в resultMap
//
// ключи из A пишутся в keysInA
//
//	и resultMap (со значением false, т.к. сами по себе ещё ничего не значат - вдруг в B их не окажется)
//
// если в B нашлось значение из keysInA, то в resultMap ему задаётся значение true
//
// если в resultMap значение true, значит значение встретилось в обоих множествах,
//
//	иначе только в A
//
// использовать одну мапу опасно - неуникальные значения из B могут быть распознаны как встреченные в A
//
// как в l1.12 - насохранял в мапу и вернул типа умный
func intersectionWithMap(a []int, b []int) []int {
	resultMap := make(map[int]bool)

	keysInA := make(map[int]struct{})

	// записать ключи A
	for _, v := range a {
		resultMap[v] = false
		keysInA[v] = struct{}{}
	}
	// записать ключи B
	for _, v := range b {
		if _, ok := keysInA[v]; ok {
			resultMap[v] = true
		}
	}

	resultSlice := make([]int, 0, len(resultMap))
	for k, v := range resultMap {
		if v {
			resultSlice = append(resultSlice, k)
		}
	}
	return resultSlice
}

func main() {
	// можно взять любые числа
	a := []int{1, 2, 3, 10}
	b := []int{2, 3, 4, 5, 10}

	fmt.Println(a, b, "пересечение", intersection(a, b))
	fmt.Println(a, b, "пересечение быстрое", intersectionWithMap(a, b))

	// [1 2 3 10] [2 3 4 5 10] пересечение [2 3 10]
	// [1 2 3 10] [2 3 4 5 10] пересечение быстрое [2 3 10]
}
