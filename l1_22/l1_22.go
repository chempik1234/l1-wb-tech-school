package main

import (
	"fmt"
	"math"
	"math/big"
)

func main() {
	// умножение, сложение и вычитание - int
	// деление - float

	// так как я все-таки получаю число через степень то получаю float, однако оно влезает в int64
	a := int64(math.Pow(2, 21))
	b := int64(math.Pow(2, 21) + 500935)

	// math/big - это число не влезает в int64
	mul := (&big.Int{}).Mul(big.NewInt(a), big.NewInt(b))

	// чтобы делить числа нужен float!
	del := float64(a) / float64(b)

	// вычитание и сложение очень простые
	add := a + b
	sub := a - b
	fmt.Printf("A=%v\tB=%v\nA*B\t=\t%v\nA/B\t=\t%v\nA+B\t=\t%v\nA-B\t=\t%v", a, b, mul, del, add, sub)

	/*
		A=2097152       B=2598087
		A*B     =       5448583348224
		A/B     =       0.8071908292524461
		A+B     =       4695239
		A-B     =       -500935
	*/
}
