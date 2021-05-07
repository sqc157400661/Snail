package main

import (
"net/http"
_ "net/http/pprof" // 第一步～
"regexp"
)

func main() {
	// 路由配置
	http.HandleFunc("/cpu", myPrint)
	_ = http.ListenAndServe("0.0.0.0:6062", nil)
}

func myPrint(writer http.ResponseWriter, request *http.Request) {
	go func() {
		for i := 0; i < 100000; i++ {
			getPhone([]string{"18505921256", "13489594009", "12759029321", "1275902932332", "127590432421", "127592433221", "127590295645", "12759045621", "12754654529321"})
		}
	}()
	_, _ = writer.Write([]byte("cpu"))
}

func getPhone(s []string) bool{
	reg := `^1([38][0-9]|14[57]|5[^4])\d{8}$`
	rgx := regexp.MustCompile(reg)
	for _, v := range s {
		//if len(v) == 11 {
			if rgx.MatchString(v) {
				return true
			}
		//}
	}
	return false
}