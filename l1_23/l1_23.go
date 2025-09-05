package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

// если не создавать новый слайс, то будет использоваться старый массив
// страшно и ужасно, целый пост в тг видел об этом в канале go-with-me
func removeAtWithMemoryLeak[T any](nums []T, index int) []T {
	copy(nums[index:], nums[index+1:])
	return nums[:len(nums)-1]
}

// здесь всё наоборот
func removeAt[T any](nums []T, index int) []T {
	result := make([]T, len(nums)-1)
	copy(result, nums[:index])
	copy(result[index:], nums[index+1:])
	return result
}

func main() {
	var pointerToOldArray *int
	var lenOldArray int
	// провернём 2 варианта удаления

	// go:noinline
	func() {
		// первый вариант аллоцирует новый массив и не меняет старый - старый можно удалить из памяти
		a := []int{1, 2, 3, 4, 5, 6}
		b := a[:]
		fmt.Printf("initial underlying array:\t%v\n", a)

		b = removeAt(b, 3)
		b = removeAt(b, 0)
		b = removeAt(b, 1)
		fmt.Printf("remove at 3, 0, 1:\t\t%v\n", b)
		fmt.Printf("underlying array:\t%v\n\n", a)

		pointerToOldArray = &a[0]
		lenOldArray = len(a)

		fmt.Printf("try to get initial array:\t\t%v\n", unsafe.Slice(pointerToOldArray, lenOldArray))
	}()

	runtime.GC()

	// да, скорее всего ничего не поменяется, ведь эта память не очищается (например memset в C) а просто поменяется как доступная(
	fmt.Printf("try to get initial array after GC:\t%v\n\n", unsafe.Slice(pointerToOldArray, lenOldArray))

	// go:noinline
	func() {
		// второй вариант продолжает использовать старый массив - он не будет удалён в любом случае!
		aa := []int{1, 2, 3, 4, 5, 6}
		bb := aa[:]
		fmt.Printf("initial underlying array:\t%v\n", aa)

		bb = removeAtWithMemoryLeak(bb, 3)
		bb = removeAtWithMemoryLeak(bb, 0)
		bb = removeAtWithMemoryLeak(bb, 1)
		fmt.Printf("MEMORY LEAK: remove at 3, 0, 1:\t%v\n", bb)

		// массив тоже поменялся ибо используется всё ещё он
		fmt.Printf("underlying array:\t\t%v\n", aa)

		pointerToOldArray = &aa[0]
		lenOldArray = len(aa)

		fmt.Printf("try to get initial array:\t\t%v\n", unsafe.Slice(pointerToOldArray, lenOldArray))
	}()

	// да, скорее всего ничего не поменяется, ведь эта память не очищается (например memset в C) а просто поменяется как доступная(
	runtime.GC()

	fmt.Printf("try to get initial array after GC:\t%v\n\n", unsafe.Slice(pointerToOldArray, lenOldArray))

	/*
		initial underlying array:       [1 2 3 4 5 6]
		remove at 3, 0, 1:              [2 5 6]
		underlying array:       [1 2 3 4 5 6]

		try to get initial array:               [1 2 3 4 5 6]
		try to get initial array after GC:      [1 2 3 4 5 6]

		initial underlying array:       [1 2 3 4 5 6]
		MEMORY LEAK: remove at 3, 0, 1: [2 5 6]
		underlying array:               [2 5 6 6 6 6]
		try to get initial array:               [2 5 6 6 6 6]
		try to get initial array after GC:      [2 5 6 6 6 6]

		мда память не corrupts красивенько по моему хотению но массив после removeAtWithMemoryLeak не уйдет в любом случае
	*/
}
