package main

import "testing"



func BenchmarkGetGoods1(b *testing.B){
	for i:=0;i<b.N ;i++  {
		getGoods1()
	}
}

func BenchmarkGetGoods2(b *testing.B){
	for i:=0;i<b.N ;i++  {
		getGoods2()
	}
}