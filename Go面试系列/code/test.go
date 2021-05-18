package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(isUnique("sdfsafs"))
	fmt.Println(isUnique("abcd"))
}

func isUnique(str string) bool {
	if strings.Count(str, "") > 3000 {
		return false
	}
	for k, v := range str {
		if v > 127 {
			return false
		}
		if strings.LastIndex(str, str[k:k+1]) != k {
			return false
		}
	}
	return true
}
