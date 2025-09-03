package main

import (
	"fmt"
	"reflect"
)

func guessType(a interface{}) string {
	// К сожалению, данный в задании v.(type) не умеет проверять, что переменная является любым каналом,
	// так что рефлексия!
	switch reflect.ValueOf(a).Kind() { // a.(type) {
	case reflect.Int:
		return "int"
	case reflect.String:
		return "string"
	case reflect.Chan:
		return "chan"
	default:
		return "unknown"
	}
}

func main() {
	// Несколько примеров значений
	var a interface{}
	var b interface{}
	var c interface{}

	a = make(chan float64)
	b = 32767
	c = "I am a string"

	fmt.Println("chan float64\t", "\trecognized as:\t", guessType(a), "\tvalue:\t", a)
	fmt.Println("int\t\t", "\trecognized as:\t", guessType(b), "\tvalue:\t", b)
	fmt.Println("string\t\t", "\trecognized as:\t", guessType(c), "\tvalue:\t", c)
}
