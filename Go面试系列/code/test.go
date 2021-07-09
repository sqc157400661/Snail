package main

import "fmt"

func main() {
	fmt.Println(checkInclusion("ab","eidbaooo"))
	fmt.Println(checkInclusion("ab","eidboaooo"))
}

func checkInclusion(s1 string, s2 string) bool {
	window,need := map[byte]int{},map[byte]int{}
	for _,v := range s1 {
		need[byte(v)]++

	}
	left,right,valid :=0,0,0
	for right < len(s2) {
		c:=s2[right]
		right++
		window[c]++
		if window[c] == need[c] {
			valid++
		}
		for valid == len(need) {
			if contain(window,need) && len(window) == len(need){
				return true
			}
			d := s2[left]
			left++
			if window[d] == need[d] {
				valid--
			}
			window[d]--
			if window[d]==0{
				delete(window,d)
			}
		}
	}
	return false
}

func contain(window,need map[byte]int) bool{
	for k,v:=range need{
		if window[k] != v {
			return false
		}
	}
	return true
}
