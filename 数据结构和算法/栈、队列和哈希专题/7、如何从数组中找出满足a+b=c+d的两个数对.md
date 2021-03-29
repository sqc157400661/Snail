## 题目描述：

给定一个数组，找出数组中是否有两个数对(a， b)和(c， d)，使得a+b=c+d，其中，a、b、c和d是不同的元素。如果有多个答案，打印任意一个即可。例如给定数组：{3， 4， 7， 10， 20， 9， 8}，可以找到两个数对 (3， 8) 和(4， 7)，使得3+8=4+7。

## 分析与解答：

最简单的方法就是使用四重遍历，对所有可能的数对判断是否满足题目要求，如果满足则打印出来，但是这种方法的时间复杂度为O(n4)，很显然不满足要求。

下面介绍另外一种方法--hash法，算法的主要思路为：以数对为单位进行遍历，在遍历过程中，把数对和数对的值存储在哈希表中（键为数对的和，值为数对），当遍历到一个键值对，如果它的和在哈希表中已经存在，那么就找到了满足条件的键值对。

下面使用Map为例给出实现代码：

```
/**
	如何从数组中找出满足a+b=c+d的两个数对
*/
package main

import (
	"fmt"
)

func main() {
	fmt.Println("如何从数组中找出满足a+b=c+d的两个数对")
	arr := []int{3,4,7,10,20,9,8}
	FindPairs(arr)
}

func FindPairs(arr []int)  {
	var reMap map[int][][2]int
	reMap = map[int][][2]int{}
	for i:=0;i<len(arr);i++{
		for j:=i+1;j<len(arr);j++{
			sum := arr[i] + arr[j]
			if reMap[sum] == nil || len(reMap[sum]) <1{
				reMap[sum] = [][2]int{{arr[i],arr[j]}}
			}else{
				reMap[sum] = append(reMap[sum], [2]int{arr[i],arr[j]})
			}
		}
	}
	for k,_ := range reMap{
		if(len(reMap[k]) >1){
			fmt.Println(reMap[k],k)
		}
	}
}

// 结果：
如何从数组中找出满足a+b=c+d的两个数对
[[3 9] [4 8]] 12
[[3 8] [4 7]] 11
[[7 10] [9 8]] 17
[[3 10] [4 9]] 13
```

## 算法性能分析：

这种方法的时间复杂度为O(n2)。因为使用了双重循环，而Map的插入与查找操作实际的时间复杂度为O(1)。