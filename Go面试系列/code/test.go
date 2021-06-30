package main

import "fmt"

func main() {
	fmt.Println(constructArr([]int{1,2,3,4,5}))
}

func constructArr(a []int) []int {
	if len(a) ==0 {
		return nil
	}
	b := make([]int,len(a))
	tmp :=1
	b[0] = 1
	// 下三角
	for i:=1; i< len(a);i++{
		b[i] = b[i-1] * a[i-1]
	}
	// 上三角
	for i:=len(a)-2;i>=0;i--{
		tmp *= a[i+1]
		b[i] *= tmp
	}
	return b
}