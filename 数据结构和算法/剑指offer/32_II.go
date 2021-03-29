package main

import "sync"

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

// 队列的实现
type queue struct {
	arr []*TreeNode
	sync.RWMutex
}

// 从队列尾部加入元素
func (q *queue) EnQueue(nums *TreeNode) {
	defer q.Unlock()
	q.Lock()
	q.arr = append(q.arr, nums)
}

// 弹出队列头部元素
func (q *queue) DeQueue() *TreeNode {
	defer q.Unlock()
	q.Lock()
	if len(q.arr) == 0 {
		return nil
	}
	re := q.arr[0]
	if len(q.arr) > 1 {
		q.arr = q.arr[1:]
	} else {
		q.arr = []*TreeNode{}
	}
	return re
}

func (q *queue) isEmpty() bool {
	return len(q.arr) == 0
}
func (q *queue) Size() int {
	return len(q.arr)
}


func levelOrder(root *TreeNode) [][]int {
	res := [][]int{}
	if root ==nil {
		return res
	}
	queue :=queue{arr:[]*TreeNode{root}}
	for !queue.isEmpty(){
		tmp :=[]int{}
		// 注意这里 每次初始化的i都是 同一级节点的数量
		for i:=queue.Size();i>0;i--{
			node := queue.DeQueue() // 访问当前节点
			tmp = append(tmp,node.Val)
			// 如果左孩子不为空则进入队列
			if node.Left !=nil{
				queue.EnQueue(node.Left)
			}
			// 如果右孩子不为空则进入队列
			if node.Right !=nil{
				queue.EnQueue(node.Right)
			}
		}
		res = append(res,tmp)

	}
	return res
}