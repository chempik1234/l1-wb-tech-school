package main

import (
	"flag"
	"fmt"
)

// reverse делает разворот строки через обмен символов с обоих краёв
// 1-й и последний, 2-й и предпоследний, и т.д.
// через XOR, ибо руны это числа
func reverse(s string) string {
	runes := []rune(s)
	for i := 0; i < len(runes)/2; i++ {
		indexFromEnd := len(runes) - i - 1
		runes[i] = runes[indexFromEnd] ^ runes[i]
		runes[indexFromEnd] = runes[indexFromEnd] ^ runes[i]
		runes[i] = runes[indexFromEnd] ^ runes[i]
	}
	return string(runes)
}

func main() {
	sFlag := flag.String("s", "главрыба", "string to reverse")
	flag.Parse()

	s := *sFlag

	fmt.Println(s, "->", reverse(s))

	/*
		go run l1_19.go -s 😊🙂🙃😊🙂🙃
		😊🙂🙃😊🙂🙃 -> 🙃🙂😊🙃🙂😊

		go run l1_19.go -s 123456789
		123456789 -> 987654321
	*/
}
