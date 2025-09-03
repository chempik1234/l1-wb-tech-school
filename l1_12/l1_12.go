package main

import "fmt"

// onlyUnique оставляет только уникальные значения
// самый банальный способ - ключи в мапе, которые быстро ищутся (проверяются на уникальность)
// насохранять ключей в мапе и вернуть слайсом
func onlyUnique(a []string) []string {
	resultMap := make(map[string]struct{})
	for _, v := range a {
		resultMap[v] = struct{}{}
	}
	resultSlice := make([]string, 0, len(resultMap))
	for k := range resultMap {
		resultSlice = append(resultSlice, k)
	}
	return resultSlice
}

func main() {
	a := []string{"cat", "cat", "dog", "cat", "tree"}

	fmt.Println(a, onlyUnique(a))
}
