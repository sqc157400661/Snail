package main

import "fmt"

func main() {
	fmt.Println(longestPalindrome("babad"))
	fmt.Println(longestPalindrome("a"))
	fmt.Println(longestPalindrome("aa"))
	fmt.Println(longestPalindrome("cbbd"))
}

func longestPalindrome(s string) string {
	res := ""
	for i := 0; i < len(s); i++ {
		// 找到以 s[i] 为中心的回文串
		s1 := palindrome(s, i, i)
		// 找到以 s[i] 和 s[i+1] 为中心的回文串
		s2:=""
		if i<len(s)-1 && s[i] == s[i+1] {
			s2 = palindrome(s, i, i+1)
		}
		res = maxLen(res, maxLen(s1, s2))
	}
	return res
}

func palindrome(s string, l, r int) string {
	n := len(s)
	// 防止索引越界
	for l >= 0 && r < n && s[l] == s[r] {
		// 向两边展开
		l--
		r++
	}
	return s[l+1:r]
}

func maxLen(s1, s2 string) string {
	if len(s1) > len(s2) {
		return s1
	}
	return s2
}
