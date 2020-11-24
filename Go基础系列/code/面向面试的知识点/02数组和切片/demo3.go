package main

import "fmt"

func main() {

	list := new([]int)

	list = append(list, 1)

	fmt.Println(list)
}