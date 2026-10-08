package main

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

func main() {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	odd := Where(nums, func(i int) bool { return i%2 == 1 })

	odd2 := slices.DeleteFunc(nums, func(i int) bool { return i%2 == 0 })
	fmt.Println("odd", odd)
	fmt.Println("odd2", odd2)

	num, err := First(nums, func(i int) bool { return i%2 == 0 })
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("first odd", num)
}

func ErrFoo() {
	_, err := Find(0)
	var validationErr *ValidationError
	if err != nil && errors.As(err, &validationErr) {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ShapesFoo() {

	circle := Circle{radius: 5}
	rect := Rectangle{length: 10, width: 5}

	CalcArea(&circle)
	CalcArea(&rect)

	a := Circle{radius: 5}
	b := Circle{radius: 6}
	c := Circle{radius: 7}

	shapes := []Shape{&a, &b, &c}
	largest := Largest(shapes)
	fmt.Println("largest", largest.Area())

	sum := SumLoop(shapes)
	fmt.Println("sum of perimeters", sum)
}

func CalcArea(s Shape) float64 {
	fmt.Println("", s.Area())
	return s.Area()
}

func boocmp(a, b Shape) int {
	return cmp.Compare(a.Area(), b.Area())
}

func Largest(shapes []Shape) Shape {
	// return slices.MaxFunc(shapes, func(a, b Shape) int {
	// 	return cmp.Compare(a.Area(), b.Area())
	// })
	return slices.MaxFunc(shapes, boocmp)
}

func SumLoop(shapes []Shape) float64 {
	var sum float64
	for _, shape := range shapes {
		sum += shape.Perimeter()
	}

	return sum
}

func slice2s() {
	//names := []string{"Foo", "Bar", "Baz"}
	names := [5]string{"Foo", "Bar", "Baz", "Alice", "Bob"} //make([]string, 3, 5)
	fmt.Println(names)

	fmt.Println("length", len(names))
	fmt.Println("capacity", cap(names))

	// [startIndex:endIndex]
	sl := names[:] //all
	fmt.Println("names[:]", sl)

	sl2 := names[1:]
	fmt.Println("[1:]", sl2)

	sl3 := names[1:3]
	fmt.Println("[1:3]", sl3)

	fmt.Println("Modifying")
	sl3[1] = strings.ToUpper(sl3[1])
	fmt.Println("[1:3]", sl3)

	fmt.Println("Checking orignal array", names)

	sl3 = append(sl3, "Charlie")
	sl3 = append(sl3, "Charlie2")
	sl3 = append(sl3, "Charlie3")
	sl3 = append(sl3, "Charlie4")
	fmt.Println("After append:", sl3)
	fmt.Println("Checking orignal array", names)
}
