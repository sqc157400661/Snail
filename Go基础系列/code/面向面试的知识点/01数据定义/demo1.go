package main

import "fmt"

/*
   下面代码是否编译通过?
*/
func myFunc(x,y int)(sum int,error){
	return x+y,nil
}

func main() {
	num, _ := myFunc(1, 2)
	fmt.Println("num = ", num)
}
