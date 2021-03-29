# 使用 Go 实现 lock-free 的队列

队列(`queue`)是非常常用的一个数据结构，它只允许在表的前端（`head`）进行出队(`dequeue`)操作，而在表的后端（`tail`）进行入队(`enqueue`)操作。和栈数据结构一样，队列是一种操作受限制的线性表。进行插入操作的端称为队尾(`tail`)，进行删除操作的端称为队头(`header`)。

在并发环境中使用队列，就必须考虑到多线程(多纤程)并发读写的问题，可能存在多个写(入队)操作线程，同时也可能存在多个线程读操作线程，在这种情况下，我们要保证数据的不丢失，不重复，而且也要保证队列的功能不变，也就是先入先出的逻辑，只要存在数据，就可以出列。

诚然，通过一个排外锁可以实现队列的并发访问。一般实现队列的时候通过指针，而且只在队头队尾操作，所以这种排外锁保护的临界区并没有很复杂的执行逻辑，临界区的处理很快，所以一般情况下通过排外锁实现队列的效率已经很高了。但是在一些情况下，通过实现 lock-free 算法，我们可以进一步提升并发队列的性能。

本文介绍 lock-free queue 算法的一些背景知识，并实现了三种并发队列，并提供了性能测试的结果。

代码库可以在github上找到: [smallnest/queue](https://github.com/smallnest/queue)。



## lock-free queue 算法

说起 lock-free queue 算法，不得不提到 Maged M. Michael 和 Michael L. Scott 1996年发表的论文 [Simple, Fast, and Practical Non-Blocking and Blocking
Concurrent Queue Algorithms](https://www.cs.rochester.edu/u/scott/papers/1996_PODC_queues.pdf)，这篇文章回顾了并发队列的一些实现以及局限性，提出了一种非常简洁的lock-free queue的实现，并且还提供了一个在特定机器比如不存在CAS指令的机器上的two-lock queue算法。这篇文章的被引用次数将近1000次。

只得一提的是, Java中的ConcurrentLinkedQueue就是基于这个算法实现的:

> This implementation employs an efficient non-blocking algorithm based on one described in Simple, Fast, and Practical Non-Blocking and Blocking Concurrent Queue Algorithms by Maged M. Michael and Michael L. Scott.

大部分lock-free的算法都是通过`CAS`操作实现的。

这篇文章提供了一个lock-free queue算法的伪代码，代码量也非常少，所以很容易通过各种编程语言实现。在这里我把伪代码列在这里:

```
structure pointer_t {ptr: pointer to node_t, count: unsigned integer}
 structure node_t {value: data type, next: pointer_t}
 structure queue_t {Head: pointer_t, Tail: pointer_t}
 
 initialize(Q: pointer to queue_t)
    node = new_node()		// Allocate a free node
    node->next.ptr = NULL	// Make it the only node in the linked list
    Q->Head.ptr = Q->Tail.ptr = node	// Both Head and Tail point to it
 
 enqueue(Q: pointer to queue_t, value: data type)
  E1:   node = new_node()	// Allocate a new node from the free list
  E2:   node->value = value	// Copy enqueued value into node
  E3:   node->next.ptr = NULL	// Set next pointer of node to NULL
  E4:   loop			// Keep trying until Enqueue is done
  E5:      tail = Q->Tail	// Read Tail.ptr and Tail.count together
  E6:      next = tail.ptr->next	// Read next ptr and count fields together
  E7:      if tail == Q->Tail	// Are tail and next consistent?
              // Was Tail pointing to the last node?
  E8:         if next.ptr == NULL
                 // Try to link node at the end of the linked list
  E9:            if CAS(&tail.ptr->next, next, <node, next.count+1>)
 E10:               break	// Enqueue is done.  Exit loop
 E11:            endif
 E12:         else		// Tail was not pointing to the last node
                 // Try to swing Tail to the next node
 E13:            CAS(&Q->Tail, tail, <next.ptr, tail.count+1>)
 E14:         endif
 E15:      endif
 E16:   endloop
        // Enqueue is done.  Try to swing Tail to the inserted node
 E17:   CAS(&Q->Tail, tail, <node, tail.count+1>)
 
 dequeue(Q: pointer to queue_t, pvalue: pointer to data type): boolean
  D1:   loop			     // Keep trying until Dequeue is done
  D2:      head = Q->Head	     // Read Head
  D3:      tail = Q->Tail	     // Read Tail
  D4:      next = head.ptr->next    // Read Head.ptr->next
  D5:      if head == Q->Head	     // Are head, tail, and next consistent?
  D6:         if head.ptr == tail.ptr // Is queue empty or Tail falling behind?
  D7:            if next.ptr == NULL  // Is queue empty?
  D8:               return FALSE      // Queue is empty, couldn't dequeue
  D9:            endif
                 // Tail is falling behind.  Try to advance it
 D10:            CAS(&Q->Tail, tail, <next.ptr, tail.count+1>)
 D11:         else		     // No need to deal with Tail
                 // Read value before CAS
                 // Otherwise, another dequeue might free the next node
 D12:            *pvalue = next.ptr->value
                 // Try to swing Head to the next node
 D13:            if CAS(&Q->Head, head, <next.ptr, head.count+1>)
 D14:               break             // Dequeue is done.  Exit loop
 D15:            endif
 D16:         endif
 D17:      endif
 D18:   endloop
 D19:   free(head.ptr)		     // It is safe now to free the old node
 D20:   return TRUE                   // Queue was not empty, dequeue succeeded
```

`initialize` 初始化一个队列，并使用一个辅助的空的节点做`header`,方便入队和出队的处理。

在入对的时候， `E1~E3`先创建一个新的节点，并把入队的数据保存在这个节点上，下一步就要插入到队尾。

`E4~E16`是一个循环，不断尝试将数据插入到队列中，在并发的情况下`CAS`可能不成功，所以胡不断尝试，并发的线程中总会有一个是成功的，所以它是一个lock-free的算法。

`E5~E6`是得到尾指针和尾指针指向的下一个节点。如果没有并发，没有并发的情况下，这里尾指针指向的下一个节点为空。但是如果在并发的情况下，在`E7`行的时候可能有别的线程已经加入了新的节点，或者先前的尾节点已经出对，所以在`E7`的实现先做一个判断，如果不满足的话重新获取。

在`E8`条件满足的情况下，说明当前获取的尾指针还是尾指针，那么在`E9`行通过`CAS`把这个节点加入到队列中，跳出循环，但是这个时候尾指针还没有改变。
否则可能在这个过程中已经有新的节点加入到队列中，那么在`E12`行，尝试把尾指针往后移动，指向新的节点。

在循环结束后，肯定已经入队，尝试把尾指针指向新插入的节点。当然这个时候可能又有新的节点加入了，导致`CAS`不成功，不过没有关系，因为这个节点已经加入了队列，只不过它已经不是尾节点了而已。更新加入的节点的逻辑会移动尾节点到最后的新加入的节点上。

在出队的时候，`D2~D4`获得头指针和尾指针，`D5`在头指针未变的情况下记一步处理，说明这个时候还没有其他出队操作。

`D6~D10`是尾指针和头指针指向的节点相同。有两种情况：1是空队列，则直接返回false,因为无数据可出列，2是新入列一个数据，还没来得及调整尾指针，那么这个时候移动一下尾指针。再重新尝试。

否则的话，`D12`先获取第一个数据，先把数据保存起来，再尝试把头指针移动到这个节点上。返回这个数据并将当前的头指针的节点数据置空，因为头指针是一个辅助节点，不需要保存数据。



## 实现

### lock-free queue



根据论文中的伪代码，我们可以使用Go语言实现一个lock-free的queue。这里指针我们使用`unsafe.Pointer`来实现，这样方便进行`CAS`操作。

```
package queue
import (
	"sync/atomic"
	"unsafe"
)
// LKQueue is a lock-free unbounded queue.
type LKQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}
type node struct {
	value interface{}
	next  unsafe.Pointer
}
// NewLKQueue returns an empty queue.
func NewLKQueue() *LKQueue {
	n := unsafe.Pointer(&node{})
	return &LKQueue{head: n, tail: n}
}
// Enqueue puts the given value v at the tail of the queue.
func (q *LKQueue) Enqueue(v interface{}) {
	n := &node{value: v}
	for {
		tail := load(&q.tail)
		next := load(&tail.next)
		if tail == load(&q.tail) { // are tail and next consistent?
			if next == nil {
				if cas(&tail.next, next, n) {
					cas(&q.tail, tail, n) // Enqueue is done.  try to swing tail to the inserted node
					return
				}
			} else { // tail was not pointing to the last node
				// try to swing Tail to the next node
				cas(&q.tail, tail, next)
			}
		}
	}
}
// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *LKQueue) Dequeue() interface{} {
	for {
		head := load(&q.head)
		tail := load(&q.tail)
		next := load(&head.next)
		if head == load(&q.head) { // are head, tail, and next consistent?
			if head == tail { // is queue empty or tail falling behind?
				if next == nil { // is queue empty?
					return nil
				}
				// tail is falling behind.  try to advance it
				cas(&q.tail, tail, next)
			} else {
				// read value before CAS otherwise another dequeue might free the next node
				v := next.value
				if cas(&q.head, head, next) {
					return v // Dequeue is done.  return
				}
			}
		}
	}
}
func load(p *unsafe.Pointer) (n *node) {
	return (*node)(atomic.LoadPointer(p))
}
func cas(p *unsafe.Pointer, old, new *node) (ok bool) {
	return atomic.CompareAndSwapPointer(
		p, unsafe.Pointer(old), unsafe.Pointer(new))
}
```

### two-lock queue

上面的lock-free queue通过`CAS`实现了高效的并发队列，同时，这篇论文还实现了一种two-lock算法，可以应用在没有原子操作的多处理器上。

```
package queue
import (
	"sync"
)
// CQueue is a concurrent unbounded queue which uses two-Lock concurrent queue qlgorithm.
type CQueue struct {
	head  *cnode
	tail  *cnode
	hlock sync.Mutex
	tlock sync.Mutex
}
type cnode struct {
	value interface{}
	next  *cnode
}
// NewCQueue returns an empty CQueue.
func NewCQueue() *CQueue {
	n := &cnode{}
	return &CQueue{head: n, tail: n}
}
// Enqueue puts the given value v at the tail of the queue.
func (q *CQueue) Enqueue(v interface{}) {
	n := &cnode{value: v}
	q.tlock.Lock()
	q.tail.next = n // Link node at the end of the linked list
	q.tail = n      // Swing Tail to node
	q.tlock.Unlock()
}
// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *CQueue) Dequeue() interface{} {
	q.hlock.Lock()
	n := q.head
	newHead := n.next
	if newHead == nil {
		q.hlock.Unlock()
		return nil
	}
	v := newHead.value
	newHead.value = nil
	q.head = newHead
	q.hlock.Unlock()
	return v
}
```

### mutex-based queue

传统的，我们可以实现一个`mutex` + slice组成的queue, 在不过分追求性能(时间+空间)的情况下实现一个简单的queue。

```
package queue
import "sync"
// SliceQueue is an unbounded queue which uses a slice as underlying.
type SliceQueue struct {
	data []interface{}
	mu   sync.Mutex
}
// NewSliceQueue returns an empty queue.
// You can give a
func NewSliceQueue(n int) (q *SliceQueue) {
	return &SliceQueue{data: make([]interface{}, n)}
}
// Enqueue puts the given value v at the tail of the queue.
func (q *SliceQueue) Enqueue(v interface{}) {
	q.mu.Lock()
	q.data = append(q.data, v)
	q.mu.Unlock()
}
// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *SliceQueue) Dequeue() interface{} {
	q.mu.Lock()
	if len(q.data) == 0 {
		q.mu.Unlock()
		return nil
	}
	v := q.data[0]
	q.data = q.data[1:]
	q.mu.Unlock()
	return v
}
```

## 性能

```
goos: darwin
goarch: amd64
pkg: github.com/smallnest/queue

BenchmarkQueue/lock-free_queue#4-4           	 8399941	       177 ns/op
BenchmarkQueue/two-lock_queue#4-4            	 7544263	       155 ns/op
BenchmarkQueue/slice-based_queue#4-4         	 6436875	       194 ns/op

BenchmarkQueue/lock-free_queue#32-4          	 8399769	       140 ns/op
BenchmarkQueue/two-lock_queue#32-4           	 7486357	       155 ns/op
BenchmarkQueue/slice-based_queue#32-4        	 4572828	       235 ns/op

BenchmarkQueue/lock-free_queue#1024-4        	 8418556	       140 ns/op
BenchmarkQueue/two-lock_queue#1024-4         	 7888488	       155 ns/op
BenchmarkQueue/slice-based_queue#1024-4      	 8902573	       218 ns/op
```

https://colobu.com/2020/08/14/lock-free-queue-in-go/