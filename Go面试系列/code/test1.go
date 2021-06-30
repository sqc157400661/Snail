package main

import "fmt"

func main() {
	var arr = []int{0, 2, 3, 5, 4}
	var arr1 = []int{0, 1, 2, 3, 5, 6}
	fmt.Println(findMiss(arr))
	fmt.Println(findMiss(arr1))
	fmt.Println(findMiss_2(arr))
	fmt.Println(findMiss_2(arr))
}

// 求和
func findMiss(arr []int) int{
	l :=len(arr)
	res := l * (l+1)/2
	sum :=0
	for i:=0;i<l;i++{
		sum += arr[i]
	}
	return res -  sum
}

// 与
func findMiss_2(arr []int) int{
	res := len(arr)
	for i:=0;i<len(arr);i++{
		res ^= arr[i]
		res ^= i
	}
	return res
}
