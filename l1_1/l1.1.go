package main

import "fmt"

// region Human

type customType struct {
	Value any
}

// Human это структура, которая имеет какие-то свои поля и методы
type Human struct {
	// age пример закрытого поля
	age int
	// Value пример открытого поля
	Value customType
}

// String это пример метода для структуры Human
func (h Human) String() string {
	return fmt.Sprintf("Human{Age:%d, Value:%v}", h.age, h.Value)
}

// GetAgeAfter это пример методы для структуры Human
func (h Human) GetAgeAfter(addYears int) int {
	return h.age + addYears
}

//endregion

//region Action

// Action встраивает Human и имеет все её поля и методы
type Action struct {
	Human
	Name               string
	PerformPeriodYears int
}

// Perform это пример метода структуры Action который не использует ничего из Human
func (a Action) Perform() {
	fmt.Println(a.Name, "performed")
}

// TimesPerformed это пример метода Action который использует поле от Human - age
//
// Даже при том что оно закрытое
func (a Action) TimesPerformed() int {
	return a.age / a.PerformPeriodYears
}

//endregion

func main() {
	a := Action{
		Human: Human{
			age:   10,
			Value: customType{Value: 1},
		},
		Name:               "parade",
		PerformPeriodYears: 3,
	}
	fmt.Printf("performed %d times\n", a.TimesPerformed())
}
