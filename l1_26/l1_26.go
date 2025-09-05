package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func allCharactersUnique(a string) bool {
	// для проверки наличия символа юзаем map
	// struct занимает меньше всего, юзаем его как и везде
	characters := make(map[rune]struct{})

	// создам переменную в scope функции чтобы её не создавать каждый раз в цикле
	var r rune
	for _, c := range a {
		// если руна нашлась - она неуникальна
		if _, ok := characters[c]; ok {
			return false
		}

		// чтобы получить руну в нижнем регистре нужно перевести её в строку, вызвать strings.ToLower
		//  и потом обратно в руну
		r, _ = utf8.DecodeRuneInString(strings.ToLower(string(c)))

		// запишем её в мапу
		characters[r] = struct{}{}
	}
	return true
}

func main() {
	for _, v := range []string{"abcd", "abCdefAaf", "aabcd", "Aa"} {
		fmt.Println(v, "->", allCharactersUnique(v))
	}

	/*
		abcd -> true
		abCdefAaf -> false
		aabcd -> false
		Aa -> false
	*/
}
