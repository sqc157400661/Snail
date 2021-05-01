## 题目描述

LRU是Least Recently Used 的缩写，它的意思是“最近最少使用”，LRU缓存就是使用这种原理实现，简单地说就是缓存一定量的数据，当超过设定的阈值时就把一些过期的数据删除掉。常用于页面置换算法，是为虚拟页式存储管理中常用的算法。下面介绍如何实现LRU缓存方案。

## 分析与解答：

我们可以使用两个数据结构实现一个LRU缓存。

1. 使用**双向链表实现的队列**，队列的最大的**容量为缓存的大小**。在使用的过程中，把**最近使用**的页面**移动到队列首**，最近**没有使用的页面**将被放在队**列尾**的位置。
2. 使用一个**哈希表**，把页号作为键，把缓存在队列中的结点的地址作为值。

当引用一个页面时，所需的页面在内存中，我们需要把这个页对应的结点移动到队列的前面。如果所需的页面不在内存中，我们将它存储在内存中。简单地说，就是将一个新结点添加到队列的前面，并在哈希表中更新相应的结点地址。如果队列是满的，那么就从队列尾部移除一个结点，并将新结点添加到队列的前面。实现代码如下：

```
/*
	实现LRU缓存方案
	LRU是Least Recently Used 的缩写，它的意思是“最近最少使用”，
	LRU缓存就是使用这种原理实现，简单地说就是缓存一定量的数据，当超过设定的阈值时就把一些过期的数据删除掉
 */
package main

import (
	"errors"
	"fmt"
	"sync"
)

/*-----------------------Set的实现，并实现了线程安全  start----------------------------*/
type Set struct {
	m map[interface{}]bool
	sync.RWMutex
}

func NewSet() *Set {
	return &Set{
		m: map[interface{}]bool{},
	}
}
func (s *Set) Add(item interface{}) {
	s.Lock()
	defer s.Unlock()
	s.m[item] = true
}
func (s *Set) Remove(item interface{}) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, item)
}
func (s *Set) Contains(item interface{}) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.m[item]
	return ok
}
func (s *Set) Len() int {
	return len(s.List())
}
func (s *Set) Clear() {
	s.RLock()
	defer s.Unlock()
	s.m = map[interface{}]bool{}
}
func (s *Set) IsEmpty() bool {
	return s.Len() == 0
}
func (s *Set) List() []interface{} {
	s.RLock()
	defer s.RUnlock()
	list := []interface{}{}
	for item := range s.m {
		list = append(list, item)
	}
	return list
}
/*-----------------------Set的实现，并实现了线程安全  end----------------------------*/

/*-----------------------Queue的实现，不是线程安全的  start----------------------------*/
// 队列定义
type SliceQueue struct {
	Arr       []int
}

// 判断队列是否是空
func (p *SliceQueue) IsEmpty() bool {
	return p.Size() == 0
}

// 获取队列的大小
func (p *SliceQueue) Size() int {
	return len(p.Arr)
}

// 获取队列首元素
func (p *SliceQueue) GetFront() interface{} {
	if p.IsEmpty() {
		return nil
	}
	return p.Arr[0]
}

// 获取队列尾元素
func (p *SliceQueue) GetBack() interface{} {
	if p.IsEmpty() {
		return nil
	}
	return p.Arr[p.Size()-1]
}

// 删除队列头元素
func (p *SliceQueue) DeQueue()  {
	if len(p.Arr) != 0{
		p.Arr = p.Arr[1:]
	}else {
		panic(errors.New("队列已经为空.."))
	}
}
// 把新元素加入队列尾
func (p *SliceQueue) EnQueue(t int) {
	p.Arr = append(p.Arr, t)
}
//把新元素加入队列首
func (p *SliceQueue) EnQueueFirst(item int) {
	newQueue := []int{item}
	p.Arr = append(newQueue, p.Arr[:]...)
}


//简单实现一个Remove
func (p *SliceQueue) Remove(item interface{}) {
	for k, v := range p.Arr {
		if v == item {
			p.Arr = append(p.Arr[:k], p.Arr[k+1:]...)
		}
	}
}
/*-----------------------Queue的实现，不是线程安全的  end----------------------------*/
// LRU结构体
type LRU struct {
	cacheSize int
	queue *SliceQueue
	hashSet *Set
}

// 判断缓存队列是否已满
func (lru *LRU) IsQueueFull() bool{
	return lru.queue.Size() == lru.cacheSize
}

// 把页号为pageNum的页也缓存到了队列中，同时也添加到Hash表中
func (lru *LRU) EnQueue(pageNum int){
	// 如果队列满了，需要先删除队尾的缓存的页面
	if lru.IsQueueFull() {
		lru.hashSet.Remove(pageNum)
	}
	// 把新缓存的节点同事添加到hash表中
	lru.hashSet.Add(pageNum)
}

/**
	当访问某一个page的时候调用这个函数，对于访问的page有2种情况
	1、如果page在缓存队列中，直接把这个节点移动到队首
	2、如果page不在缓存队列中，直接把page缓存到队首
 */
func (lru *LRU) AccessPage(pageNum int){
	// page不咋缓存队列中，把它缓存到队首
	if lru.hashSet.Contains(pageNum) {
		lru.EnQueue(pageNum)
	}else if pageNum != lru.queue.GetFront() {
		lru.queue.Remove(pageNum)
		lru.queue.EnQueueFirst(pageNum)
	}
}

func main() {
	// 创建LRU结构体
	lru := &LRU{
		cacheSize:3,
		queue: &SliceQueue{Arr: make([]int, 0)},
		hashSet:&Set{
			m: map[interface{}]bool{},
		},
	}
	lru.AccessPage(1)
	lru.AccessPage(2)
	lru.AccessPage(5)
	lru.AccessPage(2)
	fmt.Println(lru.queue)
	lru.AccessPage(4)
	fmt.Println(lru.queue)
}
```