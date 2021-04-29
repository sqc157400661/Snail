## 三足鼎立 —— GPM 到底是什么？（一）

G、P、M 是 Go 调度器的三个核心组件，各司其职。在它们精密地配合下，Go 调度器得以高效运转，这也是 Go 天然支持高并发的内在动力。今天这篇文章我们来深入理解 GPM 模型。

先看 G，取 goroutine 的首字母，主要保存 goroutine 的一些状态信息以及 CPU 的一些寄存器的值，例如 IP 寄存器，以便在轮到本 goroutine 执行时，CPU 知道要从哪一条指令处开始执行。

> 当 goroutine 被调离 CPU 时，调度器负责把 CPU 寄存器的值保存在 g 对象的成员变量之中。
>
> 当 goroutine 被调度起来运行时，调度器又负责把 g 对象的成员变量所保存的寄存器值恢复到 CPU 的寄存器。

上面这段描述来自公众号“go语言核心编程技术”的调度器系列文章，写得非常好，推荐大家去看，参考资料【阿波张调度器系列教程】可以到达原文。

本系列教程使用的代码版本是 1.9.2，来看一下 g 的源码：

本系列教程使用的代码版本是 1.9.2，来看一下 g 的源码：

```
type g struct {    // goroutine 使用的栈    stack       stack   // offset known to runtime/cgo    // 用于栈的扩张和收缩检查，抢占标志    stackguard0 uintptr // offset known to liblink    stackguard1 uintptr // offset known to liblink    _panic         *_panic // innermost panic - offset known to liblink    _defer         *_defer // innermost defer    // 当前与 g 绑定的 m    m              *m      // current m; offset known to arm liblink    // goroutine 的运行现场    sched          gobuf    syscallsp      uintptr        // if status==Gsyscall, syscallsp = sched.sp to use during gc    syscallpc      uintptr        // if status==Gsyscall, syscallpc = sched.pc to use during gc    stktopsp       uintptr        // expected sp at top of stack, to check in traceback    // wakeup 时传入的参数    param          unsafe.Pointer // passed parameter on wakeup    atomicstatus   uint32    stackLock      uint32 // sigprof/scang lock; TODO: fold in to atomicstatus    goid           int64    // g 被阻塞之后的近似时间    waitsince      int64  // approx time when the g become blocked    // g 被阻塞的原因    waitreason     string // if status==Gwaiting    // 指向全局队列里下一个 g    schedlink      guintptr    // 抢占调度标志。这个为 true 时，stackguard0 等于 stackpreempt    preempt        bool     // preemption signal, duplicates stackguard0 = stackpreempt    paniconfault   bool     // panic (instead of crash) on unexpected fault address    preemptscan    bool     // preempted g does scan for gc    gcscandone     bool     // g has scanned stack; protected by _Gscan bit in status    gcscanvalid    bool     // false at start of gc cycle, true if G has not run since last scan; TODO: remove?    throwsplit     bool     // must not split stack    raceignore     int8     // ignore race detection events    sysblocktraced bool     // StartTrace has emitted EvGoInSyscall about this goroutine    // syscall 返回之后的 cputicks，用来做 tracing    sysexitticks   int64    // cputicks when syscall has returned (for tracing)    traceseq       uint64   // trace event sequencer    tracelastp     puintptr // last P emitted an event for this goroutine    // 如果调用了 LockOsThread，那么这个 g 会绑定到某个 m 上    lockedm        *m    sig            uint32    writebuf       []byte    sigcode0       uintptr    sigcode1       uintptr    sigpc          uintptr    // 创建该 goroutine 的语句的指令地址    gopc           uintptr // pc of go statement that created this goroutine    // goroutine 函数的指令地址    startpc        uintptr // pc of goroutine function    racectx        uintptr    waiting        *sudog         // sudog structures this g is waiting on (that have a valid elem ptr); in lock order    cgoCtxt        []uintptr      // cgo traceback context    labels         unsafe.Pointer // profiler labels    // time.Sleep 缓存的定时器    timer          *timer         // cached timer for time.Sleep    gcAssistBytes int64}
```

源码中，比较重要的字段我已经作了注释，其他未作注释的与调度关系不大或者我暂时也没有理解的。

