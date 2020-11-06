package main

import "testing"

func TestMakeMap1(t *testing.T){
	MakeMap1()
}

func BenchmarkMakeMap1(b *testing.B){
	for i:=0;i<b.N ;i++  {
		MakeMap1()
	}
}

func TestMakeMap2(t *testing.T){
	MakeMap2()
}

func BenchmarkMakeMap2(b *testing.B){
	for i:=0;i<b.N ;i++  {
		MakeMap2()
	}
}