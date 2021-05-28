# Go并发编程(十) 深入理解 Channel

## 楔子

### 使用通信共享内存

**“不要通过共享内存来通信，我们应该使用通信来共享内存”** 这句话想必大家已经非常熟悉了，在官方的博客，初学时的教程，甚至是在 Go 的源码中都能看到，我们之前讲 sync 包的时候也有提到过。

无论是通过共享内存来通信还是通过通信来共享内存，最终我们应用程序都是读取的内存当中的数据，只是前者是直接读取内存的数据，而后者是通过发送消息的方式来进行同步。而通过发送消息来同步的这种方式常见的就是 **Go** 采用的 **CSP(Communication Sequential Process)** 模型以及 Erlang 采用的 Actor 模型，这两种方式都是通过通信来共享内存。
[![02_Go进阶03_blog_channel.png](D:\www\Snail\Go学习笔记\images\1610460699237-f6400aaa-34d5-4c8d-b323-27683704abd2.png)](https://mohuishou-blog-sz.oss-cn-shenzhen.aliyuncs.com/image/1610460699237-f6400aaa-34d5-4c8d-b323-27683704abd2.png)

[02_Go进阶03_blog_channel.png](https://mohuishou-blog-sz.oss-cn-shenzhen.aliyuncs.com/image/1610460699237-f6400aaa-34d5-4c8d-b323-27683704abd2.png)


大部分的语言采用的都是第一种方式直接去操作内存，然后通过互斥锁，CAS 等操作来保证并发安全。**Go 引入了 Channel 和 Goroutine 实现 CSP 模型来解耦这个操作**，这样做的好处是在 Goroutine 当中我们就不用手动去做资源的锁定与释放，同时将生产者和消费者进行了解耦，Channel 其实和消息队列很相似。而 Actor 模型和 CSP 模型都是通过发送消息来共享内存，但是它们之间最大的区别就是 Actor 模型当中并没有一个独立的 Channel 组件，而是 Actor 与 Actor 之间直接进行消息的发送与接收，每个 Actor 都有一个本地的“信箱”消息都会先发送到这个“信箱当中”。

**小结**

- 相对于互斥锁，原子操作而言 channel 是一个更高层级的抽象，使用 channel 会更加方便，心智成本也更低，同时也更不容易出错（channel 保证了并发安全），后面就会讲到，**由于 Channel 底层也是通过这些低级的同步原语实现的，所以性能上会差一些，如果有极高的性能要求时也可以用 sync 包中提供的低级同步原语**
- 使用 channel 可以帮助我们解耦生产者和消费者，可以**降低并发当中的耦合**

### happens before

在 [Week03: Go 并发编程(二) Go 内存模型](https://lailin.xyz/post/go-training-week3-go-memory-model.html) 这篇文章当中讲到同步的时候，Channel 相关部分我们特意略过了，在后面 channel 部分我们就会详细的讲到 channel 的使用以及是怎么实现的，这里先回顾一下 happens before 相关的知识点，详细可以看之前的那篇文章。

**happens before 定义:** 如果 `e1` 发生在 `e2` 之前，那么我们就说 `e2` 发生在 `e1` 之后，如果 `e1` 既不在 `e2` 前，也不在 `e2` 之后，那我们就说这俩是并发的

但是在我们进行并发编程的过程中由于编译器和 CPU 的各种优化，所以在并发执行的时候并不一定按照代码书写的顺序进行执行（在单个 Goroutine 是可以保证的），所以我们就要采用各种同步原语来保证有序，在 Go 中最常用的就是 Channel，接下来我们就进入正题吧。

**channel:**

- channel 上的发送操作总在对应的接收操作完成前发生
- 如果 channel 关闭后从中接收数据，接受者就会收到该 channel 返回的零值
- 从无缓冲的 channel 中进行的接收，要发生在对该 channel 进行的发送完成前

这些看起来会比较绕，记住这几条规则，我们接着往下走，希望可以解决你的困惑

## channel

### 基本用法

channel 的关键字为 `chan` ，使用时还需要给 channel 指定一个类型，所以完整的就是 `chan T` ，使用 `<-` 表示 channel 的数据流向，在定义变量时，我们也可以使用 `<- chan T` 、 `chan<- T` 来分别表示只读和只写的channel。
channel 的初始化采用 `make(chan T, cap)` 表示， `cap` 为可选参数，如果不填默认值为 0 表示创建了一个无缓冲的 channel。接下来我们看一个简单的例子，来了解 channel 的基本使用方法

```go
package main

import (
	"fmt"
)

// 这里只能读
func read(c <-chan int) {
	fmt.Println("read:", <-c)
}

// 这里只能写
func write(c chan<- int) {
	c <- 0
}

func main() {
	c := make(chan int)
	go read(c)
	write(c)
}
```

最后会输出 `read: 0` , 注意我们这里使用的是无缓冲的 channel，如果换成有缓存的，这里有可能就不会输出了，因为

1. **channel 上的发送操作总在对应的接收操作完成前发生，**所以在 read 还没有完成时候，write 就已经开始写入了
2. **从无缓冲的 channel 中进行的接收，要发生在对该 channel 进行的发送完成前，**如果是无缓冲的 channel， write 还没写入结束，read 就已经开始接收了，所以可以保证 read 执行，但是反过来如果有缓冲，那么 read 可能还没开始 write 就结束了，所以就有可能什么都不输出就结束了

关于有无缓冲的 channel 有两张图非常经典，基本上看完就明白了，建议阅读一下原文，在参考文献 [14] 中

#### 无缓冲 channel

如下图所示，无缓冲的 channel 会阻塞直到数据接收完成，常用于两个 goroutine 互相等待同步

![image.png](D:\www\Snail\Go学习笔记\images\1610518966242-7cf0d17a-8c64-4eb0-9459-615d10bdf56d.png)

#### 有缓冲 channel

有缓冲的 channel 如果在缓冲区未满的情况下发送是不阻塞的，在缓冲区不为空时，接收是不阻塞的

![image.png](D:\www\Snail\Go学习笔记\images\1610518977443-794150b4-3b89-4279-b18c-50ae8e334945.png)

### 源码分析

大概了解了 channel 的使用方法和原理之后我们接下来就进入稍微硬核一些的源码模式

#### 数据结构

```go
type hchan struct {
	qcount   uint           // 队列中元素总数量
	dataqsiz uint           // 循环队列的长度
	buf      unsafe.Pointer // 指向长度为 dataqsiz 的底层数组，只有在有缓冲时这个才有意义
	elemsize uint16         // 能够发送和接受的元素大小
	closed   uint32         // 是否关闭
	elemtype *_type // 元素的类型
	sendx    uint   // 当前已发送的元素在队列当中的索引位置
	recvx    uint   // 当前已接收的元素在队列当中的索引位置
	recvq    waitq  // 接收 Goroutine 链表
	sendq    waitq  // 发送 Goroutine 链表

	lock mutex // 互斥锁
}

// waitq 是一个双向链表，里面保存了 goroutine
type waitq struct {
	first *sudog
	last  *sudog
}
```

如下图所示，channel 底层其实是一个循环队列

![image.png](D:\www\Snail\Go学习笔记\images\1610547861474-f3573434-0261-4255-91dd-6e6a4a532030.png)

#### 创建

在 Go 中我们使用 `make(chan T, cap)` 来创建 channel，make 语法会在编译时，转换为 `makechan64` 和 `makechan`

```go
func makechan64(t *chantype, size int64) *hchan {
	if int64(int(size)) != size {
		panic(plainError("makechan: size out of range"))
	}

	return makechan(t, int(size))
}
```

`makechan64` 主要是做了一下检查，最终还是会调用 `makechan` ，在看 `makechan` 源码之前，我们先来看两个全局常量，接下来会用到

```go
const (
	maxAlign  = 8
	hchanSize = unsafe.Sizeof(hchan{}) + uintptr(-int(unsafe.Sizeof(hchan{}))&(maxAlign-1))
)
```

- `maxAlign` 是内存对齐的最大值，这个等于 64 位 CPU 下的 `cacheline` 的大小【有疑问应该是64byte？】
- `hchanSize` 计算 `unsafe.Sizeof(hchan{})` 最近的 8 的倍数

```go
func makechan(t *chantype, size int) *hchan {
	elem := t.elem

	// 先做一些检查
    // 元素大小不能大于等于 64k
	if elem.size >= 1<<16 {
		throw("makechan: invalid channel element type")
	}
    // 判断当前的 hchanSize 是否是 maxAlign 整数倍，并且元素的对齐大小不能大于最大对齐的大小
	if hchanSize%maxAlign != 0 || elem.align > maxAlign {
		throw("makechan: bad alignment")
	}

    // 这里计算内存是否超过限制
	mem, overflow := math.MulUintptr(elem.size, uintptr(size))
	if overflow || mem > maxAlloc-hchanSize || size < 0 {
		panic(plainError("makechan: size out of range"))
	}

	var c *hchan
	switch {
	case mem == 0: // 如果是无缓冲通道
		c = (*hchan)(mallocgc(hchanSize, nil, true)) // 为 hchan 分配内存
		c.buf = c.raceaddr() // 这个是 for data race 检测的
	case elem.ptrdata == 0: // 元素不包含指针
		c = (*hchan)(mallocgc(hchanSize+mem, nil, true)) // 为 hchan 和底层数组分配一段连续的内存地址
		c.buf = add(unsafe.Pointer(c), hchanSize)
	default: // 如果元素包含指针，分别为 hchan 和 底层数组分配内存地址
		c = new(hchan)
		c.buf = mallocgc(mem, elem, true)
	}

    // 初始化一些值
	c.elemsize = uint16(elem.size)
	c.elemtype = elem
	c.dataqsiz = uint(size)
	lockInit(&c.lock, lockRankHchan)

	return c
}
```

注释已经写得很全了，简单做个小结：

- 创建时会做一些检查
  - 元素大小不能超过 64K
  - 元素的对齐大小不能超过 maxAlign 也就是 8 字节
  - 计算出来的内存是否超过限制
- 创建时的策略
  - 如果是无缓冲的 channel，会直接给 hchan 分配内存
  - 如果是有缓冲的 channel，并且元素不包含指针，那么会为 hchan 和底层数组分配一段连续的地址
  - 如果是有缓冲的 channel，并且元素包含指针，那么会为 hchan 和底层数组分别分配地址

#### 发送数据

我们在 `x <- chan T` 进行发送数据的时候最终会被编译成 `chansend1`

```go
func chansend1(c *hchan, elem unsafe.Pointer) {
	chansend(c, elem, true, getcallerpc())
}
```

而 `chansend1` 最终还是调用了 `chansend` 主要的逻辑都在 `chansend` 上面，注意看下方源码和注释

```go
// 代码中删除了调试相关的代码
func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
    // 如果是一个 nil 值的 channel
    // 如果是非阻塞的话就直接返回
    // 如果不是，那么则调用 gopark 休眠当前 goroutine 并且抛出 panic 错误
	if c == nil {
		if !block {
			return false
		}
		gopark(nil, nil, waitReasonChanSendNilChan, traceEvGoStop, 2)
		throw("unreachable")
	}

    // fast path 如果当前是非阻塞的
    // 并且通道尚未关闭
    // 并且缓冲区已满时，直接返回
	if !block && c.closed == 0 && full(c) {
		return false
	}

    // 加锁
	lock(&c.lock)

    // 如果通道已经关闭了，直接 panic，不允许向一个已经关闭的 channel 写入数据
	if c.closed != 0 {
		unlock(&c.lock)
		panic(plainError("send on closed channel"))
	}

    // 如果当前存在等待接收数据的 goroutine 直接取出第一个，将数据传递给第一个等待的 goroutine
	if sg := c.recvq.dequeue(); sg != nil {
		// send 用于发送数据，我们后面再看
		send(c, sg, ep, func() { unlock(&c.lock) }, 3)
		return true
	}

	// 如果当前 channel 包含缓冲区，并且缓冲区没有满
	if c.qcount < c.dataqsiz {
		// 计算数组中下一个可以存放数据的地址
		qp := chanbuf(c, c.sendx)

        // 将当前的数据放到缓冲区中
		typedmemmove(c.elemtype, qp, ep)

        // 索引加一
        c.sendx++

        // 由于是循环队列，如果索引地址等于数组长度，就需要将索引移动到 0
		if c.sendx == c.dataqsiz {
			c.sendx = 0
		}

        // 当前缓存数据量加一
		c.qcount++
		unlock(&c.lock)
		return true
	}

    // 如果是非阻塞的就直接返回了，因为非阻塞发送的情况已经走完了，下面是阻塞发送的逻辑
	if !block {
		unlock(&c.lock)
		return false
	}

	// 获取发送数据的 goroutine
	gp := getg()
    // 获取 sudog 结构体，并且设置相关信息，包括当前的 channel，是否是 select 等
	mysg := acquireSudog()
	mysg.releasetime = 0
	if t0 != 0 {
		mysg.releasetime = -1
	}
	mysg.elem = ep
	mysg.waitlink = nil
	mysg.g = gp
	mysg.isSelect = false
	mysg.c = c
	gp.waiting = mysg
	gp.param = nil

    // 将 sudog 结构加入到发送的队列中
	c.sendq.enqueue(mysg)

    // 挂起当前 goroutine 等待接收 channel数据
	gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanSend, traceEvGoBlockSend, 2)

    // 保证当前数据处于活跃状态避免被回收
	KeepAlive(ep)

	// 发送者 goroutine 被唤醒，检查当前 sg 的状态
	if mysg != gp.waiting {
		throw("G waiting list is corrupted")
	}
	gp.waiting = nil
	gp.activeStackChans = false
	if gp.param == nil {
		if c.closed == 0 {
			throw("chansend: spurious wakeup")
		}
		panic(plainError("send on closed channel"))
	}
	gp.param = nil
	if mysg.releasetime > 0 {
		blockevent(mysg.releasetime-t0, 2)
	}

    // 取消 channel 绑定
	mysg.c = nil
    // 释放 sudog
	releaseSudog(mysg)
	return true
}
```

send

```go
func send(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func(), skip int) {
    // 如果 sudog 上存在数据元素，就调用 sendDirect 直接把数据拷贝到接收变量的地址上
	if sg.elem != nil {
		sendDirect(c.elemtype, sg, ep)
		sg.elem = nil
	}
	gp := sg.g
	unlockf()
	gp.param = unsafe.Pointer(sg)
	if sg.releasetime != 0 {
		sg.releasetime = cputicks()
	}

    // 调用 goready 将接受者的 Goroutine 标记为可运行状态，并把它放到发送方的所在处理器的 runnext 等待执行，下次调度时就会执行到它。
    // 注意这里不是立即执行
	goready(gp, skip+1)
}
```

**小结**
向 channel 中发送数据时大概分为两大块，检查和数据发送，而数据发送又分为三种情况

- 如果 channel 的`recvq`存在阻塞等待的接收数据的 goroutine 那么将会直接将数据发送给第一个等待的 goroutine
  - 这里会直接将数据拷贝到 `x <-ch` 接收者的变量 `x` 上
  - 然后将接收者的 Goroutine 修改为可运行状态，并把它放到发送方所在处理器的 runnext 上等待下一次调度时执行。
- 如果 channel 是有缓冲的，并且缓冲区没有满，这个时候就会把数据放到缓冲区中
- 如果 channel 的缓冲区满了，这个时候就会走阻塞发送的流程，获取到 sudog 之后将当前 Goroutine 挂起等待唤醒，唤醒后将相关的数据解绑，回收掉 sudog

#### 接收数据

在 Go 中接收 channel 数据有两种方式

- `x <- ch` 编译时会被转换为 `chanrecv1`
- `x, ok <- ch` 编译时会被转换为 `chanrecv2`

`chanrecv1` 和 `chanrecv2` 没有多大区别，只是 `chanrecv2` 比 `chanrecv1` 多了一个返回值，最终都是调用的 `chanrecv` 来实现的接收数据

```go
// selected 用于 select{} 语法中是否会选中该分支
// received 表示当前是否真正的接收到数据，用来判断 channel 是否 closed 掉了
func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
	// 和发送数据类似，先判断是否为nil，如果是 nil 并且阻塞接收就会 panic
	if c == nil {
		if !block {
			return
		}
		gopark(nil, nil, waitReasonChanReceiveNilChan, traceEvGoStop, 2)
		throw("unreachable")
	}

	// Fast path: 检查非阻塞的操作
    // empty 主要是有两种情况返回 true:
    // 1. 无缓冲channel，并且没有阻塞住发送者
    // 2. 有缓冲 channel，但是缓冲区没有数据
	if !block && empty(c) {
		// 这里判断通道是否关闭，如果是未关闭的通道说明当前还没准备好数据，直接返回
		if atomic.Load(&c.closed) == 0 {
			return
		}
		// 如果通道已经关闭了，再检查一下通道还有没有数据，如果已经没数据了，我们清理到 ep 指针中的数据并且返回
		if empty(c) {
			if ep != nil {
				typedmemclr(c.elemtype, ep)
			}
			return true, false
		}
	}

	// 上锁
	lock(&c.lock)

    // 和上面类似，如果通道已经关闭了，并且已经没数据了，我们清理到 ep 指针中的数据并且返回
	if c.closed != 0 && c.qcount == 0 {
		unlock(&c.lock)
		if ep != nil {
			typedmemclr(c.elemtype, ep)
		}
		return true, false
	}

    // 和发送类似，接收数据时也是先看一下有没有正在阻塞的等待发送数据的 Goroutine
    // 如果有的话 直接调用 recv 方法从发送者或者是缓冲区中接收数据，recv 方法后面会讲到
	if sg := c.sendq.dequeue(); sg != nil {
		recv(c, sg, ep, func() { unlock(&c.lock) }, 3)
		return true, true
	}

    // 如果 channel 的缓冲区还有数据
	if c.qcount > 0 {
		// 获取当前 channel 接收的地址
		qp := chanbuf(c, c.recvx)

        // 如果传入的指针不是 nil 直接把数据复制到对应的变量上
		if ep != nil {
			typedmemmove(c.elemtype, ep, qp)
		}
        // 清除队列中的数据，设置接受者索引并且返回
		typedmemclr(c.elemtype, qp)
		c.recvx++
		if c.recvx == c.dataqsiz {
			c.recvx = 0
		}
		c.qcount--
		unlock(&c.lock)
		return true, true
	}

    // 和发送一样剩下的就是阻塞操作了，如果是非阻塞的情况，直接返回
	if !block {
		unlock(&c.lock)
		return false, false
	}

	// 阻塞接受，和发送类似，拿到当前 Goroutine 和 sudog 并且做一些数据填充
	gp := getg()
	mysg := acquireSudog()
	mysg.releasetime = 0
	if t0 != 0 {
		mysg.releasetime = -1
	}
	mysg.elem = ep
	mysg.waitlink = nil
	gp.waiting = mysg
	mysg.g = gp
	mysg.isSelect = false
	mysg.c = c
	gp.param = nil

    // 把 sudog 放入到接收者队列当中
	c.recvq.enqueue(mysg)
    // 然后休眠当前 Goroutine 等待唤醒
	gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanReceive, traceEvGoBlockRecv, 2)

	// Goroutine 被唤醒，接收完数据，做一些数据清理的操作，释放掉 sudog 然后返回
	if mysg != gp.waiting {
		throw("G waiting list is corrupted")
	}
	gp.waiting = nil
	gp.activeStackChans = false
	if mysg.releasetime > 0 {
		blockevent(mysg.releasetime-t0, 2)
	}
	closed := gp.param == nil
	gp.param = nil
	mysg.c = nil
	releaseSudog(mysg)
	return true, !closed
}
```

**小结:** 数据接收和发送其实大同小异，也是分为检查和数据接收，数据接收又分三种情况

- 直接获取数据，如果当前有阻塞的发送者 Goroutine 走这条路
  - 如果是无缓冲 channel，直接从发送者那里把数据拷贝给接收变量
  - 如果是有缓冲 channel，并且 channel 已经满了，就先从 channel 的底层数组拷贝数据，再把阻塞的发送者 Goroutine 的数据拷贝到 channel 的循环队列中
- 从 channel 的缓冲中获取数据，有缓冲 channel 并且缓存队列有数据时走这条路
  - 直接从缓存队列中复制数据给接收变量
- 阻塞接收，剩余情况走这里
  - 和发送类似，先获取当前 Goroutine 信息，构造 sudog 加入到 channel 的 recvq 上
  - 然后休眠当前 Goroutine 等待唤醒
  - 唤醒后做一些清理工作，释放 sudog 返回

#### 关闭 channel

我们使用 `close(ch)` 来关闭 channel 最后会调用 runtime 中的 `closechan` 方法

```go
func closechan(c *hchan) {
    // 关闭 nil 的 channel 会导致 panic
	if c == nil {
		panic(plainError("close of nil channel"))
	}

    // 加锁
	lock(&c.lock)

    // 关闭已关闭的 channel 会导致 panic
	if c.closed != 0 {
		unlock(&c.lock)
		panic(plainError("close of closed channel"))
	}

	// 设置 channel 状态
	c.closed = 1

	var glist gList

	// 释放所有的接收者 Goroutine
	for {
		sg := c.recvq.dequeue()
		if sg == nil {
			break
		}
		if sg.elem != nil {
			typedmemclr(c.elemtype, sg.elem)
			sg.elem = nil
		}
		if sg.releasetime != 0 {
			sg.releasetime = cputicks()
		}
		gp := sg.g
		gp.param = nil

		glist.push(gp)
	}

	// 释放所有的发送者channel，会 panic 因为不允许向已关闭的 channel 发送数据
	for {
		sg := c.sendq.dequeue()
		if sg == nil {
			break
		}
		sg.elem = nil
		if sg.releasetime != 0 {
			sg.releasetime = cputicks()
		}
		gp := sg.g
		gp.param = nil
		if raceenabled {
			raceacquireg(gp, c.raceaddr())
		}
		glist.push(gp)
	}
	unlock(&c.lock)

	// 将所有的 Goroutine 设置为可运行状态
	for !glist.empty() {
		gp := glist.pop()
		gp.schedlink = 0
		goready(gp, 3)
	}
}
```

**小结:**

- 关闭一个 nil 的 channel 和已关闭了的 channel 都会导致 panic
- 关闭 channel 后会释放所有因为 channel 而阻塞的 Goroutine

### 使用场景

关于 channel 的使用场景在 Go 语言 101 当中已经很完善了，如果感兴趣可以看一下参考文献[5]，我这里只讲一些常见或是我觉得有趣的例子

#### 1. 通过关闭 channel 实现一对多的通知

刚刚讲到了关闭 channel 时会释放所有阻塞的 Goroutine，所以我们就可以利用这个特性来做一对多的通知，除了一对多之外我们还用了 `done` 做了多对一的通知，当然多对一这种情况还是建议直接使用 WaitGroup 即可

```go
package main

import (
	"fmt"
	"time"
)

func run(stop <-chan struct{}, done chan<- struct{}) {
	// 每一秒打印一次 hello
	for {
		select {
		case <-stop:
			fmt.Println("stop...")
			done <- struct{}{}
			return
		case <-time.After(time.Second):
			fmt.Println("hello")
		}
	}
}

func main() {
	// 一对多
	stop := make(chan struct{})
	// 多对一
	done := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		go run(stop, done)
	}

	// 5s 后退出
	time.Sleep(5 * time.Second)
	close(stop)

	for i := 0; i < 10; i++ {
		<-done
	}
}
```

#### 2. 使用 channel 做异步编程(future/promise)

其实最开始的例子就是这种情况

```go
package main

import (
	"fmt"
)

// 这里只能读
func read(c <-chan int) {
	fmt.Println("read:", <-c)
}

// 这里只能写
func write(c chan<- int) {
	c <- 0
}

func main() {
	c := make(chan int)
	go read(c)
	write(c)
}
```

#### 3. 超时控制

具体可以看案例一里面的 run 方法, 不过超时控制还是建议使用 context

```go
func run(stop <-chan struct{}, done chan<- struct{}) {
	// 每一秒打印一次 hello
	for {
		select {
		case <-stop:
			fmt.Println("stop...")
			done <- struct{}{}
			return
		case <-time.After(time.Second):
			fmt.Println("hello")
		}
	}
}
```

## 总结

这篇文章从资料收集，到源码阅读再到最终成文花了快半个月的时间了，不过写完之后就我个人而言是收获满满的，不知到能不能为屏幕前的你带来一点点启发，如果可以那就太赞了。这篇文章从最开始的理论 csp/actor 到 hanpens before 再到 channel 的基本用法，源码实现，最后讲到了一些使用场景，但是由于长度精力还是个人水平等多种限制有的地方讲解还是不够细致，具体强烈建议阅读一些参考文献里面的十几篇文章，每一个都值得细细品味。
我们下一篇文章见 👀

## 参考文献

1. [The Go Programming Language Specification - The Go Programming Language](https://golang.org/ref/spec#Channel_types)
2. [Go advanced concurrency patterns: part 3 (channels) - Blog Title](https://blogtitle.github.io/go-advanced-concurrency-patterns-part-3-channels/)
3. [Go 语言 Channel 实现原理精要 | Go 语言设计与实现](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-channel/)
4. [Go Channel 详解 | 鸟窝](https://colobu.com/2016/04/14/Golang-Channels/)
5. [通道 - Go 语言 101（通俗版 Go 白皮书）](https://gfw.go101.org/article/channel.html)
6. [通道用例大全 - Go 语言 101（通俗版 Go 白皮书）](https://gfw.go101.org/article/channel-use-cases.html)
7. [如何优雅地关闭通道 - Go 语言 101（通俗版 Go 白皮书）](https://gfw.go101.org/article/channel-closing.html)
8. [深度解密 Go 语言之 channel | qcrao](https://qcrao.com/2019/07/22/dive-into-go-channel/)
9. [为什么使用通信来共享内存 - 面向信仰编程](https://draveness.me/whys-the-design-communication-shared-memory/)
10. http://www.usingcsp.com/cspbook.pdf
11. [一文带你解密 Go 语言之通道 channel](https://mp.weixin.qq.com/s/ZXYpfLNGyej0df2zXqfnHQ)
12. [Golang 源码分析系列之 Channel 底层实现 | Tink’s Blog](https://www.cyub.vip/2020/11/04/Golang源码分析系列之Channel底层实现/)
13. [The Behavior Of Channels](https://www.ardanlabs.com/blog/2017/10/the-behavior-of-channels.html)
14. [The Nature Of Channels In Go](https://www.ardanlabs.com/blog/2014/02/the-nature-of-channels-in-go.html)

