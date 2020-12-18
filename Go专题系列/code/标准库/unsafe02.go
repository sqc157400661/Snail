package main

import "fmt"
import "unsafe"

type Rect struct {
	Width  int
	Height int
}

func main() {
	var r = Rect{50, 50}
	var width, height int
	// *Rect => Pointer => *int => int
	width =  *(*int)(unsafe.Pointer(&r))
	// *Rect => Pointer => uintptr => Pointer => *int => int
	height = *(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&r))+unsafe.Offsetof(r.Height)))
	fmt.Println(width, height)
}
