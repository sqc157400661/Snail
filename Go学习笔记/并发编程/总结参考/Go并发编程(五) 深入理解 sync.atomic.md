# Go并发编程(五) 深入理解 sync/atomic

## 回顾

在上一篇文章《[Week03: Go 并发编程(四) 深入理解 Mutex](https://lailin.xyz/post/go-training-week3-sync.html)》当中我们主要讲到了互斥锁以及读写锁的使用以及源码解析。在看源码的时候我们可以发现里面使用了很多 atomic 包的方法来保证原子，那么我们就趁热打铁接下来就随着本文来看一看 atomic 应该怎么用，以及它又是如何实现的

## 案例

上一篇文章我们在讲读写锁的时候讲到了一个配置读取的例子，这里我们使用 atomic 实现看一下

```go
// Config atomic 实现
type Config struct {
	v atomic.Value // 假设 data 就是整个 config 了
}

// Get get config data
func (c *Config) Get() []int {
	// 这里偷个懒，不要学
	return (*c.v.Load().(*[]int))
}

// Set set config data
func (c *Config) Set(n []int) {
	c.v.Store(&n)
}
```

跑一个一样的测试，可以发现 atomic 的性能又好上了许多

```
❯ go test -race -bench=.
goos: linux
goarch: amd64
pkg: github.com/mohuishou/go-training/Week03/blog/05_atomic
BenchmarkMutexConfig-4           1021684              1121 ns/op
BenchmarkRWMutexConfig-4         2604524               433 ns/op
BenchmarkConfig-4                6941658               170 ns/op
PASS
```

`atomic.Value` 这种适合配置文件这种读特别多，写特别少的场景，因为他是 COW（Copy On Write）写时复制的一种思想，COW 就是指我需要写入的时候我先把老的数据复制一份到一个新的对象，然后再写入新的值。
我们看看维基百科的描述，我觉得已经说得很清楚了

> 写入时复制（英语：Copy-on-write，简称 COW）是一种计算机程序设计领域的优化策略。其核心思想是，如果有多个调用者（callers）同时请求相同资源（如内存或磁盘上的数据存储），他们会共同获取相同的指针指向相同的资源，直到某个调用者试图修改资源的内容时，系统才会真正复制一份专用副本（private copy）给该调用者，而其他调用者所见到的最初的资源仍然保持不变。这过程对其他的调用者都是透明的。此作法主要的优点是如果调用者没有修改该资源，就不会有副本（private copy）被创建，因此多个调用者只是读取操作时可以共享同一份资源。

这种思路会有一个问题，就是可能有部分 goroutine 在使用老的对象，所以老的对象不会立即被回收，如果存在大量写入的话，会导致产生大量的副本，性能反而不一定好 。
这种方式的好处就是不用加锁，所以也不会有 goroutine 的上下文切换，并且在读取的时候大家都读取的相同的副本所以性能上回好一些。
COW 策略在 linux， redis 当中都用的很多，具体可以看一下我后面的参考文献，本文就不展开讲了。

## 源码分析

### 方法一览

如果去看文档会发现 atomic 的函数签名有很多，但是大部分都是重复的为了不同的数据类型创建了不同的签名，这就是没有泛型的坏处了，基础库会比较麻烦

1、第一类 `AddXXX` 当需要添加的值为负数的时候，做减法，正数做加法

```go
// 第一类，AddXXX，delta 为
func AddInt32(addr *int32, delta int32) (new int32)
func AddInt64(addr *int64, delta int64) (new int64)
func AddUint32(addr *uint32, delta uint32) (new uint32)
func AddUint64(addr *uint64, delta uint64) (new uint64)
func AddUintptr(addr *uintptr, delta uintptr) (new uintptr)
```

2、第二类 `CompareAndSwapXXX` CAS 操作， 会先比较传入的地址的值是否是 old，如果是的话就尝试赋新值，如果不是的话就直接返回 false，返回 true 时表示赋值成功。

```go
func CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool)
func CompareAndSwapInt64(addr *int64, old, new int64) (swapped bool)
func CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) (swapped bool)
func CompareAndSwapUint32(addr *uint32, old, new uint32) (swapped bool)
func CompareAndSwapUint64(addr *uint64, old, new uint64) (swapped bool)
func CompareAndSwapUintptr(addr *uintptr, old, new uintptr) (swapped bool)
```

3、第三类 `LoadXXX` ，从某个地址中取值

```go
func LoadInt32(addr *int32) (val int32)
func LoadInt64(addr *int64) (val int64)
func LoadPointer(addr *unsafe.Pointer) (val unsafe.Pointer)
func LoadUint32(addr *uint32) (val uint32)
func LoadUint64(addr *uint64) (val uint64)
func LoadUintptr(addr *uintptr) (val uintptr)
```

4、第四类 `StoreXXX` ，给某个地址赋值

```go
func LoadInt32(addr *int32) (val int32)
func LoadInt64(addr *int64) (val int64)
func LoadPointer(addr *unsafe.Pointer) (val unsafe.Pointer)
func LoadUint32(addr *uint32) (val uint32)
func LoadUint64(addr *uint64) (val uint64)
func LoadUintptr(addr *uintptr) (val uintptr)
```

5、第五类 `SwapXXX` ，交换两个值，并且返回老的值

```go
func SwapInt32(addr *int32, new int32) (old int32)
func SwapInt64(addr *int64, new int64) (old int64)
func SwapPointer(addr *unsafe.Pointer, new unsafe.Pointer) (old unsafe.Pointer)
func SwapUint32(addr *uint32, new uint32) (old uint32)
func SwapUint64(addr *uint64, new uint64) (old uint64)
func SwapUintptr(addr *uintptr, new uintptr) (old uintptr)
```

6、最后一类 `Value` 用于任意类型的值的 Store、Load，我们开始的案例就用到了这个，这是 1.4 版本之后引入的，签名的方法都只能作用于特定的类型，引入这个方法之后就可以用于任意类型了。

```go
type Value
func (v *Value) Load() (x interface{})
func (v *Value) Store(x interface{})
```

### CAS

在 `sync/atomic` 包中的源码除了 `Value` 之外其他的函数都是没有直接的源码的，需要去 `runtime/internal/atomic` 中找寻，这里为 `CAS` 函数为例，其他的都是大同小异

```
// bool Cas(int32 *val, int32 old, int32 new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else
//		return 0;
TEXT runtime∕internal∕atomic·Cas(SB),NOSPLIT,$0-17
	MOVQ	ptr+0(FP), BX
	MOVL	old+8(FP), AX
	MOVL	new+12(FP), CX
	LOCK
	CMPXCHGL	CX, 0(BX)
	SETEQ	ret+16(FP)
	RET
```

在注释部分写的非常清楚，这个函数主要就是先比较一下当前传入的地址的值是否和 old 值相等，如果相等，就赋值新值返回 true，如果不相等就返回 false
我们看这个具体汇编代码就可以发现，使用了 `LOCK` 来保证操作的原子性，《[Week03: Go 并发编程(二) Go 内存模型](https://lailin.xyz/post/go-training-week3-go-memory-model.html#内存重排)》提到过的一致性问题， `CMPXCHG` 指令其实就是 CPU 实现的 CAS 操作

> 关于 LOCK 指令通过查阅 intel 的手册我们可以发现，对于P6之前的处理器，LOCK 指令会总是锁总线，但是 P6 之后可能会执行“缓存锁定”，如果被锁定的内存区域被缓存在了处理器中，这个时候会通过缓存一致性来保证操作的原子性

### Value

```go
type Value struct {
	v interface{}
}
```

结构非常简单，只有一个 v 用来保存传入的值

#### Store

我们先看看 store 方法，store 方法会将值存储为 x，这里需要注意，每次传入的 x 不能为 nil，并且他们类型必须是相同的，不然会导致 panic

```go
func (v *Value) Store(x interface{}) {
	if x == nil {
		panic("sync/atomic: store of nil value into Value")
	}
    // ifaceWords 其实就是定义了一下 interface 的结构，包含 data 和 type 两部分
    // 这里 vp 是原有值
    // xp 是传入的值
	vp := (*ifaceWords)(unsafe.Pointer(v))
	xp := (*ifaceWords)(unsafe.Pointer(&x))
    // for 循环不断尝试
	for {
        // 这里先用原子方法取一下老的类型值
		typ := LoadPointer(&vp.typ)
		if typ == nil {
            // 等于 nil 就说明这是第一次 store
            // 调用 runtime 的方法禁止抢占，避免操作完成一半就被抢占了
            // 同时可以避免 GC 的时候看到 unsafe.Pointer(^uintptr(0)) 这个中间状态的值
			runtime_procPin()
			if !CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(^uintptr(0))) {
				runtime_procUnpin()
				continue
			}

			// 分别把值和类型保存下来
			StorePointer(&vp.data, xp.data)
			StorePointer(&vp.typ, xp.typ)
			runtime_procUnpin()
			return
		}

		if uintptr(typ) == ^uintptr(0) {
            // 如果判断发现这个类型是这个固定值，说明当前第一次赋值还没有完成，所以进入自旋等待
			continue
		}
		// 第一次赋值已经完成，判断新的赋值的类型和之前是否一致，如果不一致就直接 panic
		if typ != xp.typ {
			panic("sync/atomic: store of inconsistently typed value into Value")
		}
        // 保存值
		StorePointer(&vp.data, xp.data)
		return
	}
}
```

具体的逻辑都写在注释中了，这里面复杂逻辑在第一次写入，因为第一次写入的时候有两次原子写操作，所以这个时候用 typ 值作为一个判断，通过不同值判断当前所处的状态，这个在我们业务代码中其实也经常用到。然后因为引入了这个中间状态，所以又使用了 `runtime_procPin` 方法避免抢占

```go
func sync_runtime_procPin() int {
	return procPin()
}

func procPin() int {
    // 获取到当前 goroutine 的 m
	_g_ := getg()
	mp := _g_.m

    // unpin 的时候就是 locks--
	mp.locks++
	return int(mp.p.ptr().id)
}
```

#### Load

```go
func (v *Value) Load() (x interface{}) {
	vp := (*ifaceWords)(unsafe.Pointer(v))
    // 先拿到类型值
	typ := LoadPointer(&vp.typ)
    // 这个说明还没有第一次 store 或者是第一次 store 还没有完成
	if typ == nil || uintptr(typ) == ^uintptr(0) {
		// First store not yet completed.
		return nil
	}
    // 获取值
	data := LoadPointer(&vp.data)
    // 构造 x 类型
	xp := (*ifaceWords)(unsafe.Pointer(&x))
	xp.typ = typ
	xp.data = data
	return
}
```

## 实战: 实现一个“无锁”栈

```go
package main

import (
	"sync/atomic"
	"unsafe"
)

// LFStack 无锁栈
// 使用链表实现
type LFStack struct {
	head unsafe.Pointer // 栈顶
}

// Node 节点
type Node struct {
	val  int32
	next unsafe.Pointer
}

// NewLFStack NewLFStack
func NewLFStack() *LFStack {
	n := unsafe.Pointer(&Node{})
	return &LFStack{head: n}
}

// Push 入栈
func (s *LFStack) Push(v int32) {
	n := &Node{val: v}

	for {
		// 先取出栈顶
        //在进行读取value的操作的过程中,其他对此值的读写操作是可以被同时进行的,那么这个读操作很可能会读取到一个只被修改了一半的数据.因此我们要使用载入
		old := atomic.LoadPointer(&s.head)
		n.next = old
		if atomic.CompareAndSwapPointer(&s.head, old, unsafe.Pointer(n)) {
			return
		}
	}
}

// Pop 出栈，没有数据时返回 nil
func (s *LFStack) Pop() int32 {
	for {
		// 先取出栈顶
		old := atomic.LoadPointer(&s.head)
		if old == nil {
			return 0
		}

		oldNode := (*Node)(old)
		// 取出下一个节点
		next := atomic.LoadPointer(&oldNode.next)
		// 重置栈顶
		if atomic.CompareAndSwapPointer(&s.head, old, next) {
			return oldNode.val
		}
	}
}
```

这里的无锁其实只是没用互斥锁，用了原子操作，前面我们看 atomic 的源码的时候可以发现实际上在 CPU 上还是有锁的，只是我们这个锁的粒度非常小

## 总结

虽然在一些情况下 atomic 的性能要好很多，但是这个是一个 low level 的库，在实际的业务代码中最好还是使用 channel 但是我们也需要知道，在一些基础库，或者是需要极致性能的地方用上这个还是很爽的，但是使用的过程中一定要小心，不然还是会容易出 bug。

## 参考文献

1. https://pkg.go.dev/sync/atomic
2. [Go 语言标准库中 atomic.Value 的前世今生](https://blog.betacat.io/post/golang-atomic-value-exploration/)
3. [深入浅出 Go - sync/atomic 源码分析](https://xie.infoq.cn/article/562eff7a1108a7a2bc46058ca)
4. [COW 奶牛！Copy On Write 机制了解一下](https://juejin.cn/post/6844903702373859335)
5. [维基百科: COW](https://zh.wikipedia.org/wiki/寫入時複製)
6. [聊聊 CPU 的 LOCK 指令](https://albk.tech/聊聊CPU的LOCK指令.html)
7. [浅论 Lock 与 X86 Cache 一致性](https://zhuanlan.zhihu.com/p/24146167)
8. [intel 手册](https://software.intel.com/sites/default/files/managed/39/c5/325462-sdm-vol-1-2abcd-3abcd.pdf)
9. [使用 Go 实现 lock-free 的队列](https://colobu.com/2020/08/14/lock-free-queue-in-go/)