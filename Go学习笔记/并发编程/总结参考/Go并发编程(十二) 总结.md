# 总结

## 总结

最开始打算一周更新 1 - 2 篇学习笔记，这样既可以跟上课程的进度又能输出一些文章，分享一点知识，但是在写到 week03 Go 并发编程这个系列的时候发现，掉进坑里面了，毛老师虽然只讲了两节课，大概 7 小时左右，但是实际上每一个点如果要深入理解学习透还是要花费大量时间的。这也就导致了这一周的课程加上本文一共输出了十一篇文章，超过了 4W 字，但是这还没结束，还有几个常用的数据数据结构没有讲到，例如 `sync.Map` `sync.Pool` 等，这些会再接下来不定期的去更新。
[![image.png](D:\www\Snail\Go学习笔记\images\1610684635316-6d61e196-4450-40f7-99d1-0a3b5317b1bb.png)](https://mohuishou-blog-sz.oss-cn-shenzhen.aliyuncs.com/image/1610684635316-6d61e196-4450-40f7-99d1-0a3b5317b1bb.png)

[image.png](https://mohuishou-blog-sz.oss-cn-shenzhen.aliyuncs.com/image/1610684635316-6d61e196-4450-40f7-99d1-0a3b5317b1bb.png)


写完这一篇文章之后能够发现之前写的代码很多不太合理的地方，看了一圈源码之后现在也可以说在 Go 并发编程上面算是入门了。接下来我们就一起回顾一下我们之前所讲的内容。我把文章目录放在了最后，可分别点击查看。

### Goroutine

> 第一篇文章主要讲解了 Goroutine 使用的相关注意事项，或者也可以说是最佳实践

1. **请将是否异步调用的选择权交给调用者**，不然很有可能大家并不知道你在这个函数里面使用了 goroutine
2. 如果你要启动一个 goroutine 请对它负责
   1. **永远不要启动一个你无法控制它退出，或者你无法知道它何时推出的 goroutine**
   2. 还有上一篇提到的，启动 goroutine 时请加上 panic recovery 机制，避免服务直接不可用
   3. 造成 goroutine 泄漏的主要原因就是 goroutine 中造成了阻塞，并且没有外部手段控制它退出
3. **尽量避免在请求中直接启动 goroutine 来处理问题**，而应该通过启动 worker 来进行消费，这样可以避免由于请求量过大，而导致大量创建 goroutine 从而导致 oom，当然如果请求量本身非常小，那当我没说

### Go 内存模型

> 第二篇文章主要是根据 Go 官方文档中的内存模型进行阐述，后面也提到了一些 CPU 内存重排的相关知识, 主要目的是让大家知道为什么我们要使用同步原语来进行显示的同步控制

1. 编译器重排，现代编译器为了能够获取到极致的性能，可能会在编译时做一些指令重排，这就会导致有一些在单线程跑的程序在并发执行时出现一些不可预期的意外情况。
2. 内存重排，现代 CPU 大多都是多核 CPU，CPU 为了提高性能会在每个核心下设有缓存，现在一般是有三级缓存，其中一二级是 CPU 独有的，这就可能会存在在并发执行时，多个 Goroutine 在不同的 CPU 上执行看到的变量数据不一致的情况发生。
3. hanppens before，如果 `e1` 发生在 `e2` 之前，那么我们就说 `e2` 发生在 `e1` 之后，如果 `e1` 既不在 `e2` 前，也不在 `e2` 之后，那我们就说这俩是并发的
4. 机器字，对大于单个机器字的值进行读取和写入，其表现如同以不确定的顺序对多个机器字大小的值进行操作
   1. 我们现在常见的还有 32 位系统和 64 位的系统，cpu 在执行一条指令的时候对于单个机器字长的的数据的写入可以保证是原子的，对于 32 位的就是 4 个字节，对于 64 位的就是 8 个字节，对于在 32 位情况下去写入一个 8 字节的数据时就需要执行两次写入操作，这两次操作之间就没有原子性，那就可能出现先写入后半部分的数据再写入前半部分，或者是写入了一半数据然后写入失败的情况。
   2. 这也是后面我们去看源码的时候很多情况下会去做 8 字节的对齐的原因
5. init，若包 p 导入了包 q，则 q 的 init 函数会在 p 的任何函数启动前完成，函数 main.main 会在所有的 init 函数结束后启动。
   1. 不建议在生产应用依赖这个隐式的顺序
6. `go` 语句会在当前 goroutine 开始执行前启动新的 goroutine
7. goroutine 无法确保在程序中的任何事件发生之前退出
8. 解决这些问题的方法就是使用显示的同步

### data race

> 在了解了并发编程的时候为什么需要显示的同步后，这一篇文章讲述了我们如何去发现存在并发问题，也就是数据竞争的情况

1. 在编译和运行单元测试时加上 `-race` flag 就可以开启 data race 的检测
   1. 不建议在生产环境 build 的时候开启数据竞争检测，因为这会带来一定的性能损失(一般内存 5-10 倍，执行时间 2-20 倍)，当然 必须要 debug 的时候除外。
   2. 建议在执行单元测试时始终开启数据竞争的检测。
2. 总共讲解六个案例来说明哪些场景下可能会出现数据竞争，可以查看原文了解

### mutex 互斥锁

> 在了解了为什么要用同步原语，以及知道如何发现并发问题之后，我们就开始依次讲解相关的同步原语，第一篇就是 mutex

1. 首先通过一个案例引入说明了锁的基本使用方法，并且了解了互斥锁的使用原则，范围尽量小，一定要解锁并且注意顺序，小心死锁。
2. 然后讲了互斥锁的实现原理，三种模式
   1. Barging: 这种模式是为了提高吞吐量，当锁被释放时，它会唤醒第一个等待者，然后把锁给第一个等待者或者给第一个请求锁的人
   2. Handoff: 当锁释放时候，锁会一直持有直到第一个等待者准备好获取锁。它降低了吞吐量，因为锁被持有，即使另一个 goroutine 准备获取它。
   3. Spining：自旋在等待队列为空或者应用程序重度使用锁时效果不错。Parking 和 Unparking goroutines 有不低的性能成本开销，相比自旋来说要慢得多。
3. 然后我们查看分析源码了解了互斥锁和读写锁的具体实现。在 1.9 之后结合上面的三种方式，这一部分的内容比较多建议仔细阅读原文

### sync/atomic

> 互斥锁其实就是大量使用了 atomic 来实现的，所以紧接着我们就来看了 atomic 包的相关实现

1. 首先通过配置热更新的案例引入，提到了 atomic.Value 的相关使用方法
2. 然后分别介绍了 atomic 包中的各类方法的用途， `AddXXX` 等
3. 详细讲解了 CAS 的的实现原理，主要是在转换为汇编的时候使用看`LOCK`指令，然后通过查阅 Intel 手册我们知道了
   1. 对于 P6 之前的处理器，LOCK 指令会总是锁总线，但是 P6 之后可能会执行“缓存锁定”，如果被锁定的内存区域被缓存在了处理器中，这个时候会通过缓存一致性来保证操作的原子性
4. 然后详细分析了 atomic.Value 的实现源码，并且利用 atomic 实现了一个无锁栈

### sync.WaitGroup

> 在前面的案例中不止一次出现了 waitgroup 的身影，这篇文章深入分析了 waitgroup 的实现

- `WaitGroup`可以用于一个 goroutine 等待多个 goroutine 干活完成，也可以多个 goroutine 等待一个 goroutine 干活完成，是一个多对多的关系
  - 多个等待一个的典型案例是 [singleflight](https://pkg.go.dev/golang.org/x/sync/singleflight)，这个在后面将微服务可用性的时候还会再讲到，感兴趣可以看看源码
- `Add(n>0)` 方法应该在启动 goroutine 之前调用，然后在 goroution 内部调用 `Done` 方法
- `WaitGroup` 必须在 `Wait` 方法返回之后才能再次使用
- `Done` 只是 `Add` 的简单封装，所以实际上是可以通过一次加一个比较大的值减少调用，或者达到快速唤醒的目的。
- `WaitGroup` 中关于 32 位和 64 位机器的处理非常巧妙，值得学习

### errgroup

> WaitGroup 学习完了我们紧接着就学习了 errgroup，因为相对于 WaitGroup errgroup 在某些场景下更加实用

- 首先说明了 errgroup 常用的使用场景
  - 虽然 WaitGroup 已经帮我们做了很好的封装，但是仍然存在一些问题，例如如果需要返回错误，或者只要一个 goroutine 出错我们就不再等其他 goroutine 了，减少资源浪费，
- 然后分析了 errgroup 的源码，源码非常简答但是功能却很实用
  - 注意有一个坑，在后面的代码中不要把这个 ctx 当做父 context 又传给下游，因为 errgroup 取消了，这个 context 就没用了，会导致下游复用的时候出错
- 然后用 week3 的作业作为案例
  - 基于 errgroup 实现一个 http server 的启动和关闭 ，以及 linux signal 信号的注册和处理，要保证能够 一个退出，全部注销退出。

### sync.Once

> 这篇文章主要讲解了 sync.once 的使用和实现

- Once 保证了传入的函数只会执行一次，这常用在单例模式，配置文件加载，初始化这些场景下
- 但是需要注意。Once 是不能复用的，只要执行过了，再传入其他的方法也不会再执行了
- 并且 Once.Do 在执行的过程中如果 f 出现 panic，后面也不会再执行了

### context

> 这篇文章讲解的比较细致，从源码分析，到使用准则和场景以及存在的缺点都讲到了

#### 使用准则

context 包一开始就告诉了我们应该怎么用，不应该怎么用，这是应该被共同遵守的约定。

- 对 server 应用而言，传入的请求应该创建一个 context，接受
- 通过 `WithCancel` , `WithDeadline` , `WithTimeout` 创建的 Context 会同时返回一个 cancel 方法，这个方法必须要被执行，不然会导致 context 泄漏，这个可以通过执行 `go vet` 命令进行检查
- 应该将 `context.Context` 作为函数的第一个参数进行传递，参数命名一般为 `ctx` 不应该将 Context 作为字段放在结构体中。
- 不要给 context 传递 nil，如果你不知道应该传什么的时候就传递 `context.TODO()`
- 不要将函数的可选参数放在 context 当中，context 中一般只放一些全局通用的 metadata 数据，例如 tracing id 等等
- context 是并发安全的可以在多个 goroutine 中并发调用

#### 使用场景

- 超时控制
- 错误取消
- 跨 goroutine 数据同步
- 防止 goroutine 泄漏

#### 缺点

- 最显著的一个就是 context 引入需要修改函数签名，并且会病毒的式的扩散到每个函数上面，不过这个见仁见智，我看着其实还好
- 某些情况下虽然是可以做到超时返回提高用户体验，但是实际上是不会退出相关 goroutine 的，这时候可能会导致 goroutine 的泄漏，针对这个我们来看一个例子

### channel

> 这篇文章从资料收集，到源码阅读再到最终成文花了快半个月的时间，不过写完之后就我个人而言是收获满满的，不知到能不能为屏幕前的你带来一点点启发，如果可以那就太赞了。这篇文章从最开始的理论 csp/actor 到 hanpens before 再到 channel 的基本用法，源码实现，最后讲到了一些使用场景，但是由于长度精力还是个人水平等多种限制有的地方讲解还是不够细致，具体强烈建议阅读一些参考文献里面的十几篇文章，每一个都值得细细品味。

- 从 **“不要通过共享内存来通信，我们应该使用通信来共享内存”** 出发先探讨了为什么我们要这么做
- 在回到了最开始的 Go 内存模型章节故意没有讲解的部分
- 从原理到使用，说明了 channel 的基本用法
- 然后详细分析了相关源代码：实质上底层是一个循环队列
  - 数据结构
  - 如何创建
  - 发送数据
  - 接收数据
- 然后讲到了几个常用的场景
  - 通过关闭 channel 实现一对多的通知
  - 使用 channel 做异步编程(future/promise)
  - 超时控制

## Go 并发编程文章索引

我将本系列的所有文章地址都放在这里，感兴趣可以点击链接查看文章详情

1. [Week03: Go 并发编程(十一) 总结](https://lailin.xyz/post/go-training-week3-sum.html)
2. [Week03: Go 并发编程(十) 深入理解 Channel - Mohuishou](https://lailin.xyz/post/go-training-week3-channel.html)
3. [Week03: Go 并发编程(九) 深入理解 Context - Mohuishou](https://lailin.xyz/post/go-training-week3-context.html)
4. [Week03: Go 并发编程(八) 深入理解 sync.Once - Mohuishou](https://lailin.xyz/post/go-training-week3-once.html)
5. [Week03: Go 并发编程(七) 深入理解 errgroup - Mohuishou](https://lailin.xyz/post/go-training-week3-errgroup.html)
6. [Week03: Go 并发编程(六) 深入理解 WaitGroup - Mohuishou](https://lailin.xyz/post/go-training-week3-waitgroup.html)
7. [Week03: Go 并发编程(五) 深入理解 sync/atomic - Mohuishou](https://lailin.xyz/post/go-training-week3-atomic.html)
8. [Week03: Go 并发编程(四) 深入理解 Mutex - Mohuishou](https://lailin.xyz/post/go-training-week3-sync.html)
9. [Week03: Go 并发编程(三) data race - Mohuishou](https://lailin.xyz/post/go-training-week3-data-race.html)
10. [Week03: Go 并发编程(二) Go 内存模型 - Mohuishou](https://lailin.xyz/post/go-training-week3-go-memory-model.html)
11. [Week03: Go 并发编程(一) goroutine - Mohuishou](https://lailin.xyz/post/go-training-week3-goroutine.html)