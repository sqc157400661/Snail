package main

import "fmt"

func double(x *int) {
	*x += *x
	x = nil
}

func main() {
	var a = 3
	double(&a)
	fmt.Println(a) // 6

	p := &a
	double(p)
	fmt.Println(a, p == nil) // 12 false
}