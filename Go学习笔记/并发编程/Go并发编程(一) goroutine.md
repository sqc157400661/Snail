## Go并发编程(一) goroutine



## Processes and Threads

操作系统会为该应用程序创建一个进程。作为一个应用程序，**它像一个为所有资源而运行的容器**。这些**资源包括内存地址空间、文件句柄、设备和线程**。

**线程是操作系统调度的一种执行路径**，用于在处理器执行我们在函数中编写的代码。一个进程从一个线程开始，即主线程，当该线程终止时，进程终止。这是因为主线程是应用程序的原点。然后，主线程可以依次启动更多的线程，而这些线程可以启动更多的线程。

无论线程属于哪个进程，操作系统都会安排线程在可用处理器上运行。每个操作系统都有自己的算法来做出这些决定。



## Goroutines and Parallelism

Go 语言层面支持的 go 关键字，可以快速的让一个函数创建为 goroutine，我们可以认为 main 函数就是作为 goroutine 执行的。操作系统调度线程在可用处理器上运行，Go运行时调度 goroutine 在绑定到单个操作系统线程的逻辑处理器中运行（P）。即使使用这个单一的逻辑处理器和操作系统线程，也可以调度数十万 goroutine 以惊人的效率和性能并发运行。

**Concurrency is not Parallelism.**

并发不是并行。并行是指两个或多个线程同时在不同的处理器执行代码。如果将运行时配置为使用多个逻辑处理器，则调度程序将在这些逻辑处理器之间分配 goroutine，这将导致 goroutine 在不同的操作系统线程上运行。但是，要获得真正的并行性，您需要在具有多个物理处理器的计算机上运行程序。否则，goroutine 将针对单个物理处理器并发运行，即使 Go 运行时使用多个逻辑处理器。



## Keep yourself busy or do the work yourself

空的select 语句将永远阻塞。

![1620982707952](D:\www\Snail\Go学习笔记\images\1620982707952.png)

如果你的 goroutine 在从另一个 goroutine 获得结果之前无法取得进展，那么通常情况下，你自己去做这项工作比委托它( go func() )更简单。

*这通常消除了将结果从 goroutine 返回到其启动器所需的大量状态跟踪和* *chan* *操作。*

![1620983070120](D:\www\Snail\Go学习笔记\images\1620983070120.png)

## Leave concurrency to the caller

**请把是否并发的选择权交给你的调用者，而不是自己就直接悄悄的用上了 goroutine**

这两个API 有什么区别？

![1620983174477](D:\www\Snail\Go学习笔记\images\1620983174477.png)

•*将目录读取到一个* *slice* *中，然后返回整个切片，或者如果出现错误，则返回错误。这是同步调用的，ListDirectory 的调用方会阻塞，直到读取所有目录条目。根据目录的大小，这可能需要很长时间，并且可能会分配大量内存来构建目录条目名称的* *slice。*

•*ListDirectory 返回一个* *chan string，将通过该* *chan* *传递目录。当通道关闭时，这表示不再有目录。由于在 ListDirectory 返回后发生通道的填充，ListDirectory 可能内部启动* *g**oroutin**e* *来填充通道。*

当然ListDirectory chan也有自己的问题。

**ListDirectory chan 版本还有两个问题：**

•*通过使用一个关闭的通道作为不再需要处理的项目的信号，ListDirectory 无法告诉调用者通过通道返回的项目集不完整，因为中途遇到了错误。调用方无法区分空目录与完全从目录读取的错误之间的区别。这两种方法都会导致从 ListDirectory 返回的通道会立即关闭。*

•*调用者必须**持续**从通道读取，直到它关闭，因为这是调用者知道填充* *chan* 高效 *slice* *的方法快。*

![1620983578571](D:\www\Snail\Go学习笔记\images\1620983578571.png)

*filepath.WalkDir* *也是类似的模型，如果函数启动 goroutine，则必须向调用方提供显式停止该goroutine 的方法。通常，将异步执行函数的决定权交给该函数的调用方通常更容易。*



## Never start a goroutine without knowning when it will stop

在这个例子中，goroutine 泄以在 code review 快速识别出来。不幸的是，生产代码中的 goroutine 泄漏通常更难找到。我无法说明 goroutine 泄漏可能发生的所有可能方式，您可能会遇到：

![1620984309328](D:\www\Snail\Go学习笔记\images\1620984309328.png)



------

search函数是一个模拟实现，用于模拟长时间运行的操作，如数据库查询或 rpc 调用。在本例中，硬编码为200ms。定义了一个名为 process
的函数，接受字符串参数，传递给 search。对于某些应用程序，顺序调用产生的延迟可能是不可接受的。

![1620989375306](D:\www\Snail\Go学习笔记\images\1620989375306.png)

![1620989404668](D:\www\Snail\Go学习笔记\images\1620989404668.png)

![1620989417199](D:\www\Snail\Go学习笔记\images\1620989417199.png)

**Any time you start a Goroutine you must ask yourself:**

•*When will it terminate?*

•*What could prevent it from terminating?*

------

这个简单的应用程序在两个不同的端口上提供http 流量，端口8080用于应用程序流量，端口8001用于访问 /debug/pprof 端点。

![1620989703837](D:\www\Snail\Go学习笔记\images\1620989703837.png)

通过将serveApp 和 serveDebug 处理程序分解为各自的函数，我们将它们与 main.main 解耦，我们还遵循了上面的建议，并确保serveApp 和 serveDebug 将它们的并发性留给调用者。*如果 serveApp 返回，则* *main.main*
*将返回导致程序关闭**，**只能靠类似* *supervisor* *进程管理来重新启动。*

![1620989776173](D:\www\Snail\Go学习笔记\images\1620989776173.png)

然而，serveDebug是在一个单独的 goroutine 中运行的，如果它返回，那么所在的 goroutine将退出，而程序的其余部分继续运行。*由于/debug 处理程序很久以前就停止工作了，所以**其他同学**会很不高兴地发现他们无法在需要时从您的应用程序中获取统计信息。*

![1620990260566](D:\www\Snail\Go学习笔记\images\1620990260566.png)

ListenAndServer返回 nil error，最终 main.main 无法退出。log.Fatal 调用了 os.Exit，会无条件终止程序；defers 不会被调用到。

![1620990305047](D:\www\Snail\Go学习笔记\images\1620990305047.png)

*Onlyuse log.Fatal from main.main or init functions.*

