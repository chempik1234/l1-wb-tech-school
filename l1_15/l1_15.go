package main

import (
	"fmt"
	"strings"
)

var justString string

func createHugeString(length int) string {
	return strings.Repeat("؛", length)
}

func someFunc() {
	v := createHugeString(1 << 10)

	// justString = v[:100]
	//
	// да, у нас есть 100 байт и мы можем создать justString как указатель на большой массив
	//   с маленькой длиной самой строки, но
	//
	// проблемы:
	//    1. memory leak ибо указатель-то на большой массив, который в результате не будет освобождён
	//    2. нарушение длины ибо слайсинг [:100] обрезает не по количеству символов, а по размеру
	// justString = v[:100]

	// решение: копирование

	// либо байт:
	// vBytes := make([]byte, 100)
	// copy(vBytes, v[:100])
	// justString = string(vBytes)
	//
	// но это ломает UTF-8

	// либо рун:
	vRunes := []rune(v)
	croppedRunes := make([]rune, 100)
	copy(croppedRunes, vRunes[:100])
	justString = string(croppedRunes)

	// что происходит с justString?
	// в правильном случае - аллоцируется новый слайс и на его основе создаётся и присваивается строка меньшей длины
	// в оригинале - тоже присваивается строка, но на основе слайса со ссылкой на большой утекающий массив байт
}

func main() {
	someFunc()

	fmt.Println(justString)
}
