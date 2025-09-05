package main

import (
	"flag"
	"fmt"
	"strings"
)

// reverseWithSlice инвертирует порядок слов меняя их местами в слайсе - мега просто
func reverseWithSlice(s string) string {
	words := strings.Split(s, " ")
	var tmp string
	for i := 0; i < len(words)/2; i++ {
		indexFromEnd := len(words) - i - 1
		tmp = words[i]
		words[i] = words[indexFromEnd]
		words[indexFromEnd] = tmp
	}
	return strings.Join(words, " ")
}

// reverseWithConcat инвертирует порядок слов меняя их прямо в строке местами
// в задании написано "на месте" - значит на месте
func reverseWithConcat(s string) string {
	// идем 2 индексами с начала и с конца, считаем длину слов и дальше складываем кусками
	indexLeft, indexRight := 0, len(s)-1
	lenLeft, lenRight := 0, 0

	// индексы станут равными:
	// либо когда встретятся посередине среднего слова
	//
	// snow dog sun
	//       ^
	//
	// либо когда дойдут до пробела посередине
	//
	// snow dug snow sousisochkovich
	//         ^
	for indexLeft <= indexRight {
		// индексы сдвигаются друг к другу пока не наткнутся на пробел - конец слова
		leftSymbol, rightSymbol := rune(s[indexLeft]), rune(s[indexRight])
		if leftSymbol != ' ' {
			lenLeft++
			indexLeft++
		}
		if rightSymbol != ' ' {
			indexRight--
			lenRight++
		}

		// когда оба прошли по 1 слову то надо менять слова местами

		//  sun|dog car thorn|throw
		//  index: 3, len-6
		//  len: 3, 5
		if leftSymbol == ' ' && rightSymbol == ' ' {
			b := strings.Builder{}
			b.WriteString(s[:indexLeft-lenLeft])                   // до левого слова
			b.WriteString(s[indexRight+1 : indexRight+1+lenRight]) // правое слово
			b.WriteString(s[indexLeft : indexRight+1])             // между словами
			b.WriteString(s[indexLeft-lenLeft : indexLeft])        // левое слово
			b.WriteString(s[indexRight+1+lenRight:])               // после правого слова
			s = b.String()

			// indexLeft++ и indexRight-- приведут к ошибке т.к. слова разной длины
			//  один из индексов окажется дальше чем нужно,
			//  а другой вернётся на несколько символов назад и окажется посередине слова
			//
			// abs sousiska snow dog sos abs
			//             ^        ^
			//
			// abs sos snow dog sousiska abs
			//              ^       ^
			//
			// нужно считать конец каждого слова используя его новую длину:
			//
			// на месте левого слова теперь правое,
			//  поэтому от его начала надо отсчитать длину правого слова
			// на месте правого слова теперь левое,
			//  поэтому от его начала надо отсчитать длину правого слова, чтобы найти конец,
			//  и затем сдвинуться на lenLeft символов левее

			indexLeft = indexLeft - lenLeft + lenRight + 1
			indexRight = indexRight + lenRight - lenLeft - 1
			lenLeft, lenRight = 0, 0
		}
	}
	return s
}

func main() {
	sFlag := flag.String("s", "sun dog snow sousiska", "string to reverse")
	flag.Parse()

	s := *sFlag

	fmt.Println(s, "\t| with slice\t|", reverseWithSlice(s))
	fmt.Println(s, "\t| with concat\t|", reverseWithConcat(s))
}
