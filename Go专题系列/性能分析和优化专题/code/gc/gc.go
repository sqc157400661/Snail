package main

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"net/http"
	_ "net/http/pprof" // 第一步～
)
var inputData []string
func main() {
	input()
	// 路由配置
	http.HandleFunc("/mem", myPrint)
	_ = http.ListenAndServe("0.0.0.0:6062", nil)
}

func myPrint(writer http.ResponseWriter, request *http.Request) {
	_, _ = writer.Write([]byte("mem"))
}

func input(){
	for i:=0;i<100000;i++{
		inputData = append(inputData,CreateRandomString(i%10))
	}
}

func output()[]string{
	res := make([]string)
}

func CreateRandomString(len int) string  {
	var container string
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0;i < len ;i++  {
		randomInt,_ := rand.Int(rand.Reader,bigInt)
		container += string(str[randomInt.Int64()])
	}
	return container
}