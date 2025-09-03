package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sort"
)

func binarySearch(a []int, elem int) int {
	l, r := 0, len(a)-1
	for l < r {
		mid := l + (r-l)/2
		if a[mid] == elem {
			return mid
		}
		if a[mid] > elem {
			r = mid - 1
		} else if a[mid] < elem {
			l = mid + 1
		}
	}
	return -1
}

func main() {
	nFlag := flag.Int("n", 100, "amount of numbers to search in")
	eFlag := flag.Int("e", 5, "element to search for")
	flag.Parse()

	n := *nFlag
	e := *eFlag

	a := make([]int, n)
	for i := 0; i < n; i++ {
		a[i] = rand.Intn(n*2) - n
	}

	sort.Ints(a)

	fmt.Println(a)

	fmt.Println("index of", e, "is", binarySearch(a, e))

	/*
		go run l1_17.go -n 10 -e 1

		[-9 -6 -4 -4 1 4 6 7 8 8]
		index of 1 is 4
	*/
}
