package main

import "fmt"

func getLeastNumbers(arr []int, k int) []int {
	return QuickSort(arr,0,len(arr)-1,k)
}

func QuickSort(arr []int,left int,right int,k int) []int{
	if(left > right){
		return arr;
	}
	key := arr[left] //基准点
	l := left
	r := right
	for l!=r {
		// 基数在左边先移动右游标。在右边先移动左游标。
		// 移动右边，找到比基准值大的  找到后停止移动
		for arr[r] >= key && l<r{
			r--
		}
		// 移动左边 找个比基准值小的 找到后停止移动
		for arr[l] <= key && l<r{
			l++
		}

		// 左右2边都停止移动后 交换左右标记
		if(l < r){
			t := arr[l]
			arr[l] = arr[r]
			arr[r] = t
		}
	}
	// 左右2边相遇，相遇点于基准点互换
	arr[left] = arr[l];
	arr[l] = key;
	if (k == l) {
		// 正好找到最小的 k(m) 个数
		return arr[:k];
	} else if (k < l) {
		// 最小的 k 个数一定在前 m 个数中，递归划分
		return QuickSort(arr, left, l-1, k);
	} else {
		// 在右侧数组中寻找最小的 k-m 个数
		return QuickSort(arr, l+1, right, k);
	}
}



func main(){
	fmt.Println(getLeastNumbers([]int{3,1,2,0,5,6,8,9},5))
}