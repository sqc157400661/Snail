# Go并发编程(六) 深入理解 WaitGroup

在前面的几篇文章中我们或多或少都用到了 WaitGroup 来等待多个 goroutine 执行结束，今天我们来深入学习一下

## 案例

`WaitGroup` 可以解决一个 goroutine 等待多个 goroutine 同时结束的场景，这个比较常见的场景就是例如 后端 worker 启动了多个消费者干活，还有爬虫并发爬取数据，多线程下载等等。
我们这里模拟一个 worker 的例子

```go
package main

import (
	"fmt"
	"sync"
)

func worker(i int) {
	fmt.Println("worker: ", i)
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			worker(i)
		}(i)
	}
	wg.Wait()
}
```

问题: 反过来支持多个 goroutine 等待一个 goroutine 完成后再干活吗？
看我们接下来的源码分析你就知道了

## 源码分析

```go
type WaitGroup struct {
	noCopy noCopy

	// 64-bit value: high 32 bits are counter, low 32 bits are waiter count.
	// 64-bit atomic operations require 64-bit alignment, but 32-bit
	// compilers do not ensure it. So we allocate 12 bytes and then use
	// the aligned 8 bytes in them as state, and the other 4 as storage
	// for the sema.
	state1 [3]uint32
}
```

`WaitGroup` 结构十分简单，由 `nocopy` 和 `state1` 两个字段组成，其中 `nocopy` 是用来防止复制的

```go
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
```

由于嵌入了 `nocopy` 所以在执行 `go vet` 时如果检查到 `WaitGroup` 被复制了就会报错。这样可以一定程度上保证 `WaitGroup` 不被复制，对了直接 go run 是不会有错误的，所以我们代码 push 之前都会强制要求进行 lint 检查，在 ci/cd 阶段也需要先进行 lint 检查，避免出现这种类似的错误。

```
~/project/Go-000/Week03/blog/06_waitgroup/02 main*
❯ go run ./main.go

~/project/Go-000/Week03/blog/06_waitgroup/02 main*
❯ go vet .
# github.com/mohuishou/go-training/Week03/blog/06_waitgroup/02
./main.go:7:9: assignment copies lock value to wg2: sync.WaitGroup contains sync.noCopy
```

`state1` 的设计非常巧妙，这是一个是十二字节的数据，这里面主要包含两大块，counter 占用了 8 字节用于计数，sema 占用 4 字节用做信号量

为什么要这么搞呢？直接用两个字段一个表示 counter，一个表示 sema 不行么？
不行，我们看看注释里面怎么写的。

> *// 64-bit value: high 32 bits are counter, low 32 bits are waiter count.* > *// 64-bit atomic operations require 64-bit alignment, but 32-bit* > *// compilers do not ensure it. So we allocate 12 bytes and then use* > *// the aligned 8 bytes in them as state, and the other 4 as storage* > *// for the sema.*

这段话的关键点在于，在做 64 位的原子操作的时候必须要保证 64 位（8 字节）对齐，如果没有对齐的就会有问题，但是 32 位的编译器并不能保证 64 位对齐所以这里用一个 12 字节的 state1 字段来存储这两个状态，然后根据是否 8 字节对齐选择不同的保存方式。
[![02_Go进阶03_blog_waitgroup.drawio.svg](D:\www\Snail\Go学习笔记\images\1609085423413-88a8f508-0269-4cf9-9474-9f78b78a53ea.svg)](https://mohuishou-blog-sz.oss-cn-shenzhen.aliyuncs.com/image/1609085423413-88a8f508-0269-4cf9-9474-9f78b78a53ea.svg)

[02_Go进阶03_blog_waitgroup.drawio.svg](https://mohuishou-blog-sz.oss-cn-shenzhen.aliyuncs.com/image/1609085423413-88a8f508-0269-4cf9-9474-9f78b78a53ea.svg)


这个操作巧妙在哪里呢？

- 如果是 64 位的机器那肯定是 8 字节对齐了的，所以是上面第一种方式
- 如果在 32 位的机器上
  - 如果恰好 8 字节对齐了，那么也是第一种方式取前面的 8 字节数据
  - 如果是没有对齐，但是 32 位 4 字节是对齐了的，所以我们只需要后移四个字节，那么就 8 字节对齐了，所以是第二种方式

所以通过 sema 信号量这四个字节的位置不同，保证了 counter 这个字段无论在 32 位还是 64 为机器上都是 8 字节对齐的，后续做 64 位原子操作的时候就没问题了。
这个实现是在 `state` 方法实现的

```go
func (wg *WaitGroup) state() (statep *uint64, semap *uint32) {
	if uintptr(unsafe.Pointer(&wg.state1))%8 == 0 {
		return (*uint64)(unsafe.Pointer(&wg.state1)), &wg.state1[2]
	} else {
		return (*uint64)(unsafe.Pointer(&wg.state1[1])), &wg.state1[0]
	}
}
```

`state` 方法返回 counter 和信号量，通过 `uintptr(unsafe.Pointer(&wg.state1))%8 == 0` 来判断是否 8 字节对齐

### Add

```go
func (wg *WaitGroup) Add(delta int) {
    // 先从 state 当中把数据和信号量取出来
	statep, semap := wg.state()

    // 在 waiter 上加上 delta 值
	state := atomic.AddUint64(statep, uint64(delta)<<32)
    // 取出当前的 counter
	v := int32(state >> 32)
    // 取出当前的 waiter，正在等待 goroutine 数量
	w := uint32(state)

    // counter 不能为负数
	if v < 0 {
		panic("sync: negative WaitGroup counter")
	}

    // 这里属于防御性编程
    // w != 0 说明现在已经有 goroutine 在等待中，说明已经调用了 Wait() 方法
    // 这时候 delta > 0 && v == int32(delta) 说明在调用了 Wait() 方法之后又想加入新的等待者
    // 这种操作是不允许的
	if w != 0 && delta > 0 && v == int32(delta) {
		panic("sync: WaitGroup misuse: Add called concurrently with Wait")
	}
    // 如果当前没有人在等待就直接返回，并且 counter > 0
	if v > 0 || w == 0 {
		return
	}

    // 这里也是防御 主要避免并发调用 add 和 wait
	if *statep != state {
		panic("sync: WaitGroup misuse: Add called concurrently with Wait")
	}

	// 唤醒所有 waiter，看到这里就回答了上面的问题了
	*statep = 0
	for ; w != 0; w-- {
		runtime_Semrelease(semap, false, 0)
	}
}
```

### Wait

wait 主要就是等待其他的 goroutine 完事之后唤醒

```go
func (wg *WaitGroup) Wait() {
	// 先从 state 当中把数据和信号量的地址取出来
    statep, semap := wg.state()

	for {
     	// 这里去除 counter 和 waiter 的数据
		state := atomic.LoadUint64(statep)
		v := int32(state >> 32)
		w := uint32(state)

        // counter = 0 说明没有在等的，直接返回就行
        if v == 0 {
			// Counter is 0, no need to wait.
			return
		}

		// waiter + 1，调用一次就多一个等待者，然后休眠当前 goroutine 等待被唤醒
		if atomic.CompareAndSwapUint64(statep, state, state+1) {
			runtime_Semacquire(semap)
			if *statep != 0 {
				panic("sync: WaitGroup is reused before previous Wait has returned")
			}
			return
		}
	}
}
```

### Done

这个只是 add 的简单封装

```go
func (wg *WaitGroup) Done() {
	wg.Add(-1)
}
```

## 总结

- WaitGroup可以用于一个 goroutine 等待多个 goroutine 干活完成，也可以多个 goroutine 等待一个 goroutine 干活完成，是一个多对多的关系
  - 多个等待一个的典型案例是 [singleflight](https://pkg.go.dev/golang.org/x/sync/singleflight)，这个在后面将微服务可用性的时候还会再讲到，感兴趣可以看看源码
- `Add(n>0)` 方法应该在启动 goroutine 之前调用，然后在 goroution 内部调用 `Done` 方法
- `WaitGroup` 必须在 `Wait` 方法返回之后才能再次使用
- `Done` 只是 `Add` 的简单封装，所以实际上是可以通过一次加一个比较大的值减少调用，或者达到快速唤醒的目的。

## 参考文献

1. https://pkg.go.dev/sync 官网文档，必读
2. [Go 语言设计与实现-6.2 同步原语与锁](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-sync-primitives/) 这本书值得一看
3. [Go by Example: WaitGroups](https://gobyexample.com/waitgroups)
4. https://golang.org/issues/8005#issuecomment-190753527
5. [Golang 源码系列 sync.waitgroup 源码剖析](https://juejin.cn/post/6893019249263001613)