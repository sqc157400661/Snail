package main

import (
	"fmt"
	"strings"
	"unicode"
)

func main() {
	fmt.Println(replaceBlank("45 123"))
	fmt.Println(replaceBlank("ae c d"))
}

func replaceBlank(s string) (string, bool) {
	if len([]rune(s)) > 1000 {
		return s, false
	}
	for _, v := range s {
		if string(v) != " " && unicode.IsLetter(v) == false {
			return s, false
		}
	}
	return strings.Replace(s, " ", "%20", -1), true
}
