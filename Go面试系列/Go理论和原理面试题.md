## Golang 理论

### Goroutine调度策略

https://mp.weixin.qq.com/s?__biz=MzU1OTg5NDkzOA==&mid=2247483801&idx=1&sn=ef7f872afccf148661cbd5a3d3b5b0a2&scene=19#wechat_redirect

```
原文： 第三章 Goroutine调度策略（16）
在调度器概述一节我们提到过，所谓的goroutine调度，是指程序代码按照一定的算法
在适当的时候挑选出合适的goroutine并放到CPU上去运行的过程。这句话揭示了调度
系统需要解决的三大核心问题：
 调度时机：什么时候会发生调度？
 调度策略：使用什么策略来挑选下一个进入运行的goroutine？
 切换机制：如何把挑选出来的goroutine放到CPU上运行？
对这三大问题的解决构成了调度器的所有工作，因而我们对调度器的分析也必将围绕着
它们所展开。
第二章我们已经详细的分析了调度器的初始化以及goroutine的切换机制，本章将重点
讨论调度器如何挑选下一个goroutine出来运行的策略问题，而剩下的与调度时机相关
的内容我们将在第4～6章进行全面的分析。
```
### 再探schedule函数

```
在讨论main goroutine的调度时我们已经⻅过schedule函数，因为当时我们的主要关注
点在于main goroutine是如何被调度到CPU上运行的，所以并未对schedule函数如何挑
选下一个goroutine出来运行做深入的分析，现在是重新回到schedule函数详细分析其
调度策略的时候了。
runtime/proc.go : 2467
```
```
5 }
```
```
1 // One round of scheduler: find a runnable goroutine and execute it.
```
(^2) // Never returns.
(^3) func schedule() {
(^4) _g_ := getg() //_g_ = m.g0
(^56) ......
(^78) var gp *g
(^109) ......
11
12 if gp == nil {
13 // Check the global runnable queue once in a while to ensure fairness.
(^14) // Otherwise two goroutines can completely occupy the local runqueue
(^15) // by constantly respawning each other.


```
schedule函数分三步分别从各运行队列中寻找可运行的goroutine：
 第一步，从全局运行队列中寻找goroutine。为了保证调度的公平性，每个工作线程
每经过61次调度就需要优先尝试从全局运行队列中找出一个goroutine来运行，这样
才能保证位于全局运行队列中的goroutine得到调度的机会。全局运行队列是所有工
作线程都可以访问的，所以在访问它之前需要加锁。
 第二步，从工作线程本地运行队列中寻找goroutine。如果不需要或不能从全局运行
队列中获取到goroutine则从本地运行队列中获取。
 第三步，从其它工作线程的运行队列中偷取goroutine。如果上一步也没有找到需要
运行的goroutine，则调用findrunnable从其他工作线程的运行队列中偷取
goroutine，findrunnable函数在偷取之前会再次尝试从全局运行队列和当前线程的
本地运行队列中查找需要运行的goroutine。
下面我们先来看如何从全局运行队列中获取goroutine。
```
### 从全局运行队列中获取goroutine

```
//为了保证调度的公平性，每个工作线程每进行 61 次调度就需要优先从全局运行队列中
获取goroutine出来运行，
```
16

```
//因为如果只调度本地运行队列中的goroutine，则全局运行队列中的goroutine有可
能得不到运行
```
17

(^18) if _g_.m.p.ptr().schedtick%61 == 0 && sched.runqsize > 0 {
(^19) lock(&sched.lock) //所有工作线程都能访问全局运行队列，所以需要加锁
gp = globrunqget(_g_.m.p.ptr(), 1) //从全局运行队列中获取 1 个
goroutine
20
21 unlock(&sched.lock)
22 }
(^23) }
(^24) if gp == nil {
(^25) //从与m关联的p的本地运行队列中获取goroutine
(^26) gp, inheritTime = runqget(_g_.m.p.ptr())
(^27) if gp != nil && _g_.m.spinning {
(^28) throw("schedule: spinning with local work")
29 }
30 }
31 if gp == nil {
(^32) //如果从本地运行队列和全局运行队列都没有找到需要运行的goroutine，
//则调用findrunnable函数从其它工作线程的运行队列中偷取，如果偷取不到，则
当前工作线程进入睡眠，
33
(^34) //直到获取到需要运行的goroutine之后findrunnable函数才会返回。
(^35) gp, inheritTime = findrunnable() // blocks until work is available
36 }
3738 ......
3940 //当前运行的是runtime的代码，函数调用栈使用的是g0的栈空间
(^41) //调用execte切换到gp的代码和栈空间去运行
(^42) execute(gp, inheritTime)
(^43) }


```
从全局运行队列中获取可运行的goroutine是通过globrunqget函数来完成的，该函数的
第一个参数是与当前工作线程绑定的p，第二个参数max表示最多可以从全局队列中拿多
少个g到当前工作线程的本地运行队列中来。
runtime/proc.go : 4663
```
```
globrunqget函数首先会根据全局运行队列中goroutine的数量，函数参数max以及_p_
的本地队列的容量计算出到底应该拿多少个goroutine，然后把第一个g结构体对象通过
返回值的方式返回给调用函数，其它的则通过runqput函数放入当前工作线程的本地运
行队列。这段代码值得一提的是，计算应该从全局运行队列中拿走多少个goroutine时
根据p的数量（gomaxprocs）做了负载均衡。
如果没有从全局运行队列中获取到goroutine，那么接下来就在工作线程的本地运行队
列中寻找需要运行的goroutine。
```
### 从工作线程本地运行队列中获取goroutine

```
从代码上来看，工作线程的本地运行队列其实分为两个部分，一部分是由p的runq、
runqhead和runqtail这三个成员组成的一个无锁循环队列，该队列最多可包含256个
```
```
1 // Try get a batch of G's from the global runnable queue.
2 // Sched must be locked.
```
(^3) func globrunqget(_p_ *p, max int32) *g {
(^4) if sched.runqsize == 0 { //全局运行队列为空
(^5) return nil
(^6) }
(^78) //根据p的数量平分全局运行队列中的goroutines
(^9) n := sched.runqsize / gomaxprocs + 1
if n > sched.runqsize { //上面计算n的方法可能导致n大于全局运行队列中的
goroutine数量
10
11 n = sched.runqsize
(^12) }
(^13) if max > 0 && n > max {
(^14) n = max //最多取max个goroutine
(^15) }
(^16) if n > int32(len(_p_.runq)) / 2 {
17 n = int32(len(_p_.runq)) / 2 //最多只能取本地队列容量的一半
18 }
1920 sched.runqsize -= n
(^2122) //直接通过函数返回gp，其它的goroutines通过runqput放入本地运行队列
(^23) gp := sched.runq.pop() //pop从全局运行队列的队列头取
(^24) n--
(^25) for ; n > 0; n-- {
(^26) gp1 := sched.runq.pop() //从全局运行队列中取出一个goroutine
(^27) runqput(_p_, gp1, false) //放入本地运行队列
28 }
29 return gp
30 }


```
goroutine；另一部分是p的runnext成员，它是一个指向g结构体对象的指针，它最多只
包含一个goroutine。
从本地运行队列中寻找goroutine是通过 runqget 函数完成的，寻找时，代码首先查看
runnext 成员是否为空，如果不为空则返回runnext所指的goroutine，并把runnext成
员清零，如果runnext为空，则继续从循环队列中查找goroutine。
runtime/proc.go : 4825
```
```
这里首先需要注意的是不管是从runnext还是从循环队列中拿取goroutine都使用了cas
操作，这里的cas操作是必需的，因为可能有其他工作线程此时此刻也正在访问这两个
成员，从这里偷取可运行的goroutine。
其次，代码中对runqhead的操作使用了 atomic.LoadAcq 和atomic.CasRel ，它们分别
提供了load-acquire 和cas-release 语义。
对于atomic.LoadAcq来说，其语义主要包含如下几条：
 原子读取，也就是说不管代码运行在哪种平台，保证在读取过程中不会有其它线程
对该变量进行写入；
```
```
1 // Get g from local runnable queue.
```
(^2) // If inheritTime is true, gp should inherit the remaining time in the
(^3) // current time slice. Otherwise, it should start a new time slice.
(^4) // Executed only by the owner P.
(^5) func runqget(_p_ *p) (gp *g, inheritTime bool) {
(^6) // If there's a runnext, it's the next G to run.
7 //从runnext成员中获取goroutine
8 for {
(^9) //查看runnext成员是否为空，不为空则返回该goroutine
(^10) next := _p_.runnext
(^11) if next == 0 {
(^12) break
(^13) }
(^14) if _p_.runnext.cas(next, 0) {
15 return next.ptr(), true
16 }
17 }
(^1819) //从循环队列中获取goroutine
(^20) for {
h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize
with other consumers
21
(^22) t := _p_.runqtail
(^23) if t == h {
24 return nil, false
25 }
(^26) gp := _p_.runq[h%uint32(len(_p_.runq))].ptr()
if atomic.CasRel(&_p_.runqhead, h, h+1) { // cas-release, commits
consume
27
(^28) return gp, false
(^29) }
(^30) }
31 }


 位于 atomic.LoadAcq 之后的代码，对内存的读取和写入必须在atomic.LoadAcq 读
取完成后才能执行，编译器和CPU都不能打乱这个顺序；
 当前线程执行atomic.LoadAcq 时可以读取到其它线程最近一次通过 atomic.CasRel
对同一个变量写入的值，与此同时，位于atomic.LoadAcq 之后的代码，不管读取哪
个内存地址中的值，都可以读取到其它线程中位于atomic.CasRel（对同一个变量操
作）之前的代码最近一次对内存的写入。
对于atomic.CasRel来说，其语义主要包含如下几条：
 原子的执行比较并交换的操作；
 位于 atomic.CasRel 之前的代码，对内存的读取和写入必须在atomic.CasRel 对内存
的写入之前完成，编译器和CPU都不能打乱这个顺序；
 线程执行 atomic.CasRel 完成后其它线程通过 atomic.LoadAcq 读取同一个变量可以
读到最新的值，与此同时，位于 atomic.CasRel 之前的代码对内存写入的值，可以
被其它线程中位于 atomic.LoadAcq（对同一个变量操作）之后的代码读取到。
因为可能有多个线程会并发的修改和读取runqhead ，以及需要依靠runqhead的值来读
取runq数组的元素，所以需要使用atomic.LoadAcq和atomic.CasRel来保证上述语义。
我们可能会问，为什么读取p的runqtail成员不需要使用atomic.LoadAcq或
atomic.load？因为runqtail不会被其它线程修改，只会被当前工作线程修改，此时没有
人修改它，所以也就不需要使用原子相关的操作。
最后，由 p的 runq 、runqhead 和runqtail 这三个成员组成的这个无锁循环队列非
常精妙，我们会在后面的章节对这个循环队列进行分析。

### CAS操作与ABA问题

我们知道使用cas操作需要特别注意ABA的问题，那么runqget函数这两个使用cas的地
方会不会有问题呢？答案是这两个地方都不会有ABA的问题。原因分析如下：
首先来看对runnext的cas操作。只有跟_p_绑定的当前工作线程才会去修改runnext为一
个非0值，其它线程只会把runnext的值从一个非0值修改为0值，然而跟_p_绑定的当前
工作线程正在此处执行代码，所以在当前工作线程读取到值A之后，不可能有线程修改
其值为B(0)之后再修改回A。
再来看对runq的cas操作。当前工作线程操作的是_p_的本地队列，只有跟_p_绑定在一
起的当前工作线程才会因为往该队列里面添加goroutine而去修改runqtail，而其它工作
线程不会往该队列里面添加goroutine，也就不会去修改runqtail，它们只会修改
runqhead，所以，当我们这个工作线程从runqhead读取到值A之后，其它工作线程也就
不可能修改runqhead的值为B之后再第二次把它修改为值A（因为runqtail在这段时间之
内不可能被修改，runqhead的值也就无法越过runqtail再回绕到A值），也就是说，代码
从逻辑上已经杜绝了引发ABA的条件。
到此，我们已经分析完工作线程从全局运行队列和本地运行队列获取goroutine的代
码，由于篇幅的限制，我们下一节再来分析从其它工作线程的运行队列偷取goroutine
的流程。

### goroutine简介

goroutine是Go语言实现的用户态线程，主要用来解决操作系统线程太“重”的问题，所
谓的太重，主要表现在以下两个方面：


####  创建和切换太重：操作系统线程的创建和切换都需要进入内核，而进入内核所消耗

#### 的性能代价比较高，开销较大；

####  内存使用太重：一方面，为了尽量避免极端情况下操作系统线程栈的溢出，内核在

#### 创建操作系统线程时默认会为其分配一个较大的栈内存（虚拟地址空间，内核并不

#### 会一开始就分配这么多的物理内存），然而在绝大多数情况下，系统线程远远用不

#### 了这么多内存，这导致了浪费；另一方面，栈内存空间一旦创建和初始化完成之后

#### 其大小就不能再有变化，这决定了在某些特殊场景下系统线程栈还是有溢出的⻛

#### 险。

```
而相对的，用户态的goroutine则轻量得多：
 goroutine是用户态线程，其创建和切换都在用户代码中完成而无需进入操作系统内
核，所以其开销要远远小于系统线程的创建和切换；
 goroutine启动时默认栈大小只有2k，这在多数情况下已经够用了，即使不够用，
goroutine的栈也会自动扩大，同时，如果栈太大了过于浪费它还能自动收缩，这样
既没有栈溢出的⻛险，也不会造成栈内存空间的大量浪费。
正是因为Go语言中实现了如此轻量级的线程，才使得我们在Go程序中，可以轻易的创
建成千上万甚至上百万的goroutine出来并发的执行任务而不用太担心性能和内存等问
题。
注意： 为了避免混淆，从现在开始，后面出现的所有的线程一词均是指操作系统线程，
而goroutine我们不再称之为什么什么线程而是直接使用goroutine这个词。
```
### 线程模型与调度器

```
第一章讨论操作系统线程调度的时候我们曾经提到过，goroutine建立在操作系统线程
基础之上，它与操作系统线程之间实现了一个多对多(M:N)的两级线程模型。
这里的 M:N 是指M个goroutine运行在N个操作系统线程之上，内核负责对这N个操作系
统线程进行调度，而这N个系统线程又负责对这M个goroutine进行调度和运行。
所谓的对goroutine的调度，是指程序代码按照一定的算法在适当的时候挑选出合适的
goroutine并放到CPU上去运行的过程，这些负责对goroutine进行调度的程序代码我们
称之为goroutine调度器。用极度简化了的伪代码来描述goroutine调度器的工作流程大
概是下面这个样子：
1 // 程序启动时的初始化代码
```
(^2) ......
(^3) for i := 0; i < N; i++ { // 创建N个操作系统线程执行schedule函数
(^4) create_os_thread(schedule) // 创建一个操作系统线程执行schedule函数
(^5) }
(^67) //schedule函数实现调度逻辑
8 func schedule() {
9 for { //调度循环
10 // 根据某种算法从M个goroutine中找出一个需要运行的goroutine
(^11) g := find_a_runnable_goroutine_from_M_goroutines()
(^12) run_g(g) // CPU运行该goroutine，直到需要调度其它goroutine才返回
(^13) save_status_of_g(g) // 保存goroutine的状态，主要是寄存器的值
(^14) }
(^15) }


#### 这段伪代码表达的意思是，程序运行起来之后创建了N个由内核调度的操作系统线程

```
（为了方便描述，我们称这些系统线程为工作线程）去执行shedule函数，而schedule
函数在一个调度循环中反复从M个goroutine中挑选出一个需要运行的goroutine并跳转
到该goroutine去运行，直到需要调度其它goroutine时才返回到schedule函数中通过
save_status_of_g保存刚刚正在运行的goroutine的状态然后再次去寻找下一个
goroutine。
需要强调的是，这段伪代码对goroutine的调度代码做了高度的抽象、修改和简化处
理，放在这里只是为了帮助我们从宏观上了解goroutine的两级调度模型，具体的实现
原理和细节将从本章开始进行全面介绍。
```
### 重要的结构体

#### 下面介绍的这些结构体中的字段非常多，牵涉到的细节也很庞杂，光是看这些结构体的

#### 定义我们没有必要也无法真正理解它们的用途，所以在这里我们只需要大概了解一下就

#### 行了，看不懂记不住都没有关系，随着后面对代码逐步深入的分析，我们也必将会对这

#### 些结构体有越来越清晰的认识。为了节省篇幅，下面各结构体的定义略去了跟调度器无

```
关的成员。另外，这些结构体的定义全部位于Go语言的源代码路径下的runtime/runtim
e2.go文件之中。
```
### stack结构体

```
stack结构体主要用来记录goroutine所使用的栈的信息，包括栈顶和栈底位置：
```
### gobuf结构体

```
gobuf结构体用于保存goroutine的调度信息，主要包括CPU的几个寄存器的值：
```
```
1 // Stack describes a Go execution stack.
```
(^2) // The bounds of the stack are exactly [lo, hi),
(^3) // with no implicit data structures on either side.
(^4) //用于记录goroutine使用的栈的起始和结束位置
(^5) type stack struct {
6 lo uintptr // 栈顶，指向内存低地址
7 hi uintptr // 栈底，指向内存高地址
8 }
1 type gobuf struct {
2 // The offsets of sp, pc, and g are known to (hard-coded in) libmach.
3 //
(^4) // ctxt is unusual with respect to GC: it may be a
(^5) // heap-allocated funcval, so GC needs to track it, but it
(^6) // needs to be set and cleared from assembly, where it's
(^7) // difficult to have write barriers. However, ctxt is really a
(^8) // saved, live register, and we only ever exchange it between
9 // the real register and the gobuf. Hence, we treat it as a
10 // root during stack scanning, which means assembly that saves


### g结构体

```
g结构体用于代表一个goroutine，该结构体保存了goroutine的所有信息，包括栈，
gobuf结构体和其它的一些状态信息：
```
11 // and restores it doesn't need write barriers. It's still
12 // typed as a pointer so that any other writes from Go get

(^13) // write barriers.
(^14) sp uintptr // 保存CPU的rsp寄存器的值
(^15) pc uintptr // 保存CPU的rip寄存器的值
(^16) g guintptr // 记录当前这个gobuf对象属于哪个goroutine
(^17) ctxt unsafe.Pointer
18
19 // 保存系统调用的返回值，因为从系统调用返回之后如果p被其它工作线程抢占，
// 则这个goroutine会被放入全局运行队列被其它工作线程调度，其它线程需要知道系统
调用的返回值。
20
(^21) ret sys.Uintreg
(^22) lr uintptr
(^23)
(^24) // 保存CPU的rip寄存器的值
(^25) bp uintptr // for GOEXPERIMENT=framepointer
26 }
1 // 前文所说的g结构体，它代表了一个goroutine
2 type g struct {
3 // Stack parameters.
(^4) // stack describes the actual stack memory: [stack.lo, stack.hi).
// stackguard0 is the stack pointer compared in the Go stack growth
prologue.
5
// It is stack.lo+StackGuard normally, but can be StackPreempt to
trigger a preemption.
6
// stackguard1 is the stack pointer compared in the C stack growth
prologue.
7
8 // It is stack.lo+StackGuard on g0 and gsignal stacks.
// It is ~0 on other goroutine stacks, to trigger a call to morestackc
(and crash).
9
(^10)
(^11) // 记录该goroutine使用的栈
(^12) stack stack // offset known to runtime/cgo
(^13) // 下面两个成员用于栈溢出检查，实现栈的自动伸缩，抢占调度也会用到stackguard0
14 stackguard0 uintptr // offset known to liblink
15 stackguard1 uintptr // offset known to liblink
1617 ......
(^18)
(^19) // 此goroutine正在被哪个工作线程执行
(^20) m *m // current m; offset known to arm liblink
(^21) // 保存调度信息，主要是几个寄存器的值
(^22) sched gobuf


### m结构体

```
m结构体用来代表工作线程，它保存了m自身使用的栈信息，当前正在运行的goroutine
以及与m绑定的p等信息，详⻅下面定义中的注释：
```
23
24 ......

(^25) // schedlink字段指向全局运行队列中的下一个g，
(^26) //所有位于全局运行队列中的g形成一个链表
(^27) schedlink guintptr
(^2829) ......
(^30) // 抢占调度标志，如果需要抢占调度，设置preempt为true
preempt bool // preemption signal, duplicates stackguard0
= stackpreempt
31
3233 ......
(^34) }
1 type m struct {
(^2) // g0主要用来记录工作线程使用的栈信息，在执行调度代码时需要使用这个栈
(^3) // 执行用户goroutine代码时，使用用户goroutine自己的栈，调度时会发生栈的切换
(^4) g0 *g // goroutine with scheduling stack
(^56) // 通过TLS实现m结构体对象与工作线程之间的绑定
tls [6]uintptr // thread-local storage (for x86 extern
register)
7
8 mstartfn func()
9 // 指向工作线程正在运行的goroutine的g结构体对象
(^10) curg *g // current running goroutine
(^11)
(^12) // 记录与当前工作线程绑定的p结构体对象
p puintptr // attached p for executing go code (nil if not
executing go code)
13
14 nextp puintptr
oldp puintptr // the p that was attached before executing a
syscall
15
(^16)
// spinning状态：表示当前工作线程正在试图从其它工作线程的本地运行队列偷取
goroutine
17
spinning bool // m is out of work and is actively looking for
work
18
(^19) blocked bool // m is blocked on a note
20
21 // 没有goroutine需要运行时，工作线程睡眠在这个park成员上，
22 // 其它线程通过这个park唤醒该工作线程
(^23) park note
(^24) // 记录所有工作线程的一个链表
(^25) alllink *m // on allm
(^26) schedlink muintptr
(^2728) // Linux平台thread的值就是操作系统线程ID


### p结构体

```
p结构体用于保存工作线程执行go代码时所必需的资源，比如goroutine的运行队列，内
存分配用到的缓存等等。
```
### schedt结构体

```
schedt结构体用来保存调度器的状态信息和goroutine的全局运行队列：
```
29 thread uintptr // thread handle
30 freelink *m // on sched.freem

(^3132) ......
(^33) }
1 type p struct {
(^2) lock mutex
(^34) status uint32 // one of pidle/prunning/...
(^5) link puintptr
6 schedtick uint32 // incremented on every scheduler call
7 syscalltick uint32 // incremented on every system call
8 sysmontick sysmontick // last tick observed by sysmon
(^9) m muintptr // back-link to associated m (nil if idle)
(^1011) ......
(^1213) // Queue of runnable goroutines. Accessed without lock.
(^14) //本地goroutine运行队列
(^15) runqhead uint32 // 队列头
(^16) runqtail uint32 // 队列尾
17 runq [256]guintptr //使用数组实现的循环队列
18 // runnext, if non-nil, is a runnable G that was ready'd by
19 // the current G and should be run next instead of what's in
(^20) // runq if there's time remaining in the running G's time
(^21) // slice. It will inherit the time left in the current time
(^22) // slice. If a set of goroutines is locked in a
(^23) // communicate-and-wait pattern, this schedules that set as a
(^24) // unit and eliminates the (potentially large) scheduling
25 // latency that otherwise arises from adding the ready'd
26 // goroutines to the end of the run queue.
27 runnext guintptr
(^2829) // Available G's (status == Gdead)
(^30) gFree struct {
(^31) gList
(^32) n int32
(^33) }
(^3435) ......
36 }
1 type schedt struct {


### 重要的全局变量

```
// accessed atomically. keep at top to ensure alignment on 32-bit
systems.
```
```
2
```
(^3) goidgen uint64
(^4) lastpoll uint64
(^56) lock mutex
(^78) // When increasing nmidle, nmidlelocked, nmsys, or nmfreed, be
(^9) // sure to call checkdead().
1011 // 由空闲的工作线程组成链表
12 midle muintptr // idle m's waiting for work
13 // 空闲的工作线程的数量
(^14) nmidle int32 // number of idle m's waiting for work
(^15) nmidlelocked int32 // number of locked m's waiting for work
mnext int64 // number of m's that have been created and next
M ID
16
(^17) // 最多只能创建maxmcount个工作线程
(^18) maxmcount int32 // maximum number of m's allowed (or die)
19 nmsys int32 // number of system m's not counted for deadlock
20 nmfreed int64 // cumulative number of freed m's
2122 ngsys uint32 // number of system goroutines; updated atomically
(^2324) // 由空闲的p结构体对象组成的链表
(^25) pidle puintptr // idle p's
(^26) // 空闲的p结构体对象的数量
(^27) npidle uint32
nmspinning uint32 // See "Worker thread parking/unparking" comment in
proc.go.
28
2930 // Global runnable queue.
31 // goroutine全局运行队列
(^32) runq gQueue
(^33) runqsize int32
(^3435) ......
(^3637) // Global cache of dead G's.
(^38) // gFree是所有已经退出的goroutine对应的g结构体对象组成的链表
(^39) // 用于缓存g结构体对象，避免每次创建goroutine时都重新分配内存
40 gFree struct {
41 lock mutex
42 stack gList // Gs with stacks
(^43) noStack gList // Gs without stacks
(^44) n int32
(^45) }
(^46)
(^47) ......
48 }
1 allgs []*g // 保存所有的g
2 allm *m // 所有的m构成的一个链表，包括下面的m0


```
在程序初始化时，这些全变量都会被初始化为0值，指针会被初始化为nil指针，切片初
始化为nil切片，int被初始化为数字0，结构体的所有成员变量按其本类型初始化为其类
型的0值。所以程序刚启动时allgs，allm和allp都不包含任何g,m和p。
```

