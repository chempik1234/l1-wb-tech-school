package main

import (
	"fmt"
	"math"
)

type Point struct {
	x, y float64
}

func NewPoint(x, y float64) *Point {
	return &Point{x, y}
}

func (p *Point) String() string {
	return fmt.Sprintf("(%f, %f)", p.x, p.y)
}

func (p *Point) Distance(q *Point) float64 {
	// теорема пифагора, логично
	return math.Sqrt(math.Pow(p.x-q.x, 2) + math.Pow(p.y-q.y, 2))
}

func main() {
	p1 := NewPoint(1, 2395.48195)
	p2 := NewPoint(85914124125.581857, 1481.48175)

	fmt.Printf("point 1:\t%s\npoint 2:\t%s\ndistance:\t%v\n", p1.String(), p2.String(), p1.Distance(p2))
}
