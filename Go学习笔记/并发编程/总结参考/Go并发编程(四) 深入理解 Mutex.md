## Go并发编程(四) 深入理解 Mutex

# 回顾

前面几篇文章当中我们都反复提到了 goroutine 建是简单的，但是我们仍然要小心，习惯总会不经意间的导致我们写出很多 bug 对于语言规范没有定义的内容我们不要做任何假设。我们需要通过同步语义控制他们的执行顺序，关于之前的内容可以看前面的三篇文章：

- [Week03: Go 并发编程(一) goroutine](https://lailin.xyz/post/go-training-week3-goroutine.html)
- [Week03: Go 并发编程(二) Go 内存模型](https://lailin.xyz/post/go-training-week3-go-memory-model.html)
- [Week03: Go 并发编程(三) data race](https://lailin.xyz/post/go-training-week3-data-race.html)

接下来的几篇文章就让我们我们一起来了解一下 sync 包相关的一些用法，以及部分实现原理，当然这里说是 sync 包，实际上包含了三个包分别是: sync, sync/atomic, golang.org/x/sync/errgroup

这些包提供了一些基础的同步语义，但是在实际的并发编程当中，我们应该使用 channel 来进行同步控制。“Share memory by communicating; don’t communicate by sharing memory.”

# Mutex

## 案例

我们先来看一下上一篇文章说到的例子应该怎么改

```go
var mu sync.Mutex

func main() {
	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go routine(i)
	}
	wg.Wait()
	fmt.Printf("Final Counter: %d\n", counter)
}

func routine(id int) {
	for i := 0; i < 2; i++ {
		mu.Lock()
		counter++
		mu.Unlock()
	}
	wg.Done()
}
```

这里主要的目的就是为了保护我们临界区的数据，通过锁来进行保证。锁的使用非常的简单，但是还是有几个需要注意的点

- 锁的范围要尽量的小，不要搞很多大锁
- 用锁一定要解锁，小心产生死锁

## 实现原理

我们来看一下在 Go 中锁是怎么实现的

### 锁的实现模式[5]

- **Barging**: 这种模式是为了提高吞吐量，当锁被释放时，它会唤醒第一个等待者，然后把锁给第一个等待者或者给第一个请求锁的人

![1_B1atM-b6GPDS0_Q_TPEUBw.png](D:\www\Snail\Go学习笔记\images\1608967840701-d8c54ee4-964b-49d0-8fab-6f98dd777fd2.png)

- **Handoff:** 当锁释放时候，锁会一直持有直到第一个等待者准备好获取锁。它降低了吞吐量，因为锁被持有，即使另一个 goroutine 准备获取它。**这种模式可以解决公平性的问题，因为在 Barging 模式下可能会存在被唤醒的 goroutine 永远也获取不到锁的情况，毕竟一直在 cpu 上跑着的 goroutine 没有上下文切换会更快一些。缺点就是性能会相对差一些**

  ![image.png](D:\www\Snail\Go学习笔记\images\1608967902210-d4c2937f-56fd-49e8-a5a3-903bec31e6fc.png)

- **Spining：**自旋在等待队列为空或者应用程序重度使用锁时效果不错。Parking 和 Unparking goroutines 有不低的性能成本开销，相比自旋来说要慢得多。***但是自旋是有成本的，所以在 go 的实现中进入自旋的条件十分的苛刻。***

  ![image.png](D:\www\Snail\Go学习笔记\images\1608967913891-eb4cf780-6ffd-4a2d-a3d5-8e05d7e3fce3.png)

### Go Mutex 实现原理

我们先来看一下在 Go 中具体是怎么实现的，我们先讲原理再看源码，避免看的云里雾里的。**

#### 加锁

如下图所示，Go 在 1.15 的版本中锁的实现结合上面提到的三种模式，调用 Lock 方法的时候。

1. 首先如果当前锁处于初始化状态就直接用 CAS 方法尝试获取锁，这是**_ Fast Path_**

2. 如果失败就进入 ***Slow Path***

   1. 会首先判断当前能不能进入自旋状态，如果可以就进入自旋，最多自旋 4 次

   2. 自旋完成之后，就会去计算当前的锁的状态

   3. 然后尝试通过 CAS 获取锁

   4. 如果没有获取到就调用 `runtime_SemacquireMutex` 方法休眠当前 goroutine 并且尝试获取信号量

   5. goroutine 被唤醒之后会先判断当前是否处在饥饿状态，（如果当前 goroutine 超过 1ms 都没有获取到锁就会进饥饿模式） 1. 如果处在饥饿状态就会获得互斥锁，如果等待队列中只存在当前 Goroutine，互斥锁还会从饥饿模式中退出 1. 如果不在，就会设置唤醒和饥饿标记、重置迭代次数并重新执行获取锁的循环

      > CAS 方法在这里指的是 `atomic.CompareAndSwapInt32(addr, old, new) bool` 方法，这个方法会先比较传入的地址的值是否是 old，如果是的话就尝试赋新值，如果不是的话就直接返回 false，返回 true 时表示赋值成功
      > 饥饿模式是 Go 1.9 版本之后引入的优化，用于解决公平性的问题[10]

      ![02_Go进阶03_blog_sync.drawio.svg](D:\www\Snail\Go学习笔记\images\1608970759375-09d8cda7-77ac-48d3-b2f3-b8890e927bd4.svg)

      

#### 解锁

解锁的流程相对于加锁简单许多
[![02_Go进阶03_blog_sync.drawio.svg](D:\www\Snail\Go学习笔记\images\1608978117259-455cf28e-aa1e-46cf-8fd6-6040ed6c0a7a.svg)](https://mohuishou-blog-sz.oss-cn-shenzhen.aliyuncs.com/image/1608978117259-455cf28e-aa1e-46cf-8fd6-6040ed6c0a7a.svg)





## 源码分析

### Mutex 基本结构

知道其中的原理之后，我们再来看看源码分析

```go
type Mutex struct {
	state int32
	sema  uint32
}
```

`Mutex` 结构体由 `state` `sema` 两个 4 字节成员组成，其中 `state` 表示了当前锁的状态， `sema` 是用于控制锁的信号量
[![02_Go进阶03_blog_sync.drawio.svg](D:\www\Snail\Go学习笔记\images\1608972241012-8c0fe8e2-b1c8-4696-a9c4-454e11753e0f.svg)](https://mohuishou-blog-sz.oss-cn-shenzhen.aliyuncs.com/image/1608972241012-8c0fe8e2-b1c8-4696-a9c4-454e11753e0f.svg)


`state`字段的最低三位表示三种状态，分别是`mutexLocked``mutexWoken``mutexStarving`，剩下的用于统计当前在等待锁的 goroutine 数量

- `mutexLocked` 表示是否处于锁定状态
- `mutexWoken` 表示是否处于唤醒状态
- `mutexStarving` 表示是否处于饥饿状态

### 加锁

回味一下上面看到的流程图，我们来看看互斥锁是如何加锁的

```go
func (m *Mutex) Lock() {
	// Fast path: grab unlocked mutex.
	if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
		return
	}
	// Slow path (outlined so that the fast path can be inlined)
	m.lockSlow()
}
```

- 当我们调用 `Lock` 方法的时候，会先尝试走 Fast Path，也就是如果当前互斥锁如果处于未加锁的状态，尝试加锁，只要加锁成功就直接返回
- 否则的话就进入 `slow path`

```go
func (m *Mutex) lockSlow() {
	var waitStartTime int64 // 等待时间
	starving := false // 是否处于饥饿状态
	awoke := false // 是否处于唤醒状态
	iter := 0 // 自旋迭代次数
	old := m.state
	for {
		// Don't spin in starvation mode, ownership is handed off to waiters
		// so we won't be able to acquire the mutex anyway.
		if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
			// Active spinning makes sense.
			// Try to set mutexWoken flag to inform Unlock
			// to not wake other blocked goroutines.
			if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
				atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
				awoke = true
			}
			runtime_doSpin()
			iter++
			old = m.state
			continue
		}
```

在 `lockSlow` 方法中我们可以看到，有一个大的 for 循环，不断的尝试去获取互斥锁，在循环的内部，第一步就是判断能否自旋状态。
进入自旋状态的判断比较苛刻，具体需要满足什么条件呢？ `runtime_canSpin` 源码见下方

- 当前互斥锁的状态是非饥饿状态，并且已经被锁定了
- 自旋次数不超过 4 次
- cpu 个数大于一，必须要是多核 cpu
- 当前正在执行当中，并且队列空闲的 p 的个数大于等于一

```go
// Active spinning for sync.Mutex.
//go:linkname sync_runtime_canSpin sync.runtime_canSpin
//go:nosplit
func sync_runtime_canSpin(i int) bool {
	if i >= active_spin || ncpu <= 1 || gomaxprocs <= int32(sched.npidle+sched.nmspinning)+1 {
		return false
	}
	if p := getg().m.p.ptr(); !runqempty(p) {
		return false
	}
	return true
}
```

如果可以进入自旋状态之后就会调用 `runtime_doSpin` 方法进入自旋， `doSpin` 方法会调用 `procyield(30)` 执行三十次 `PAUSE` 指令

```
TEXT runtime·procyield(SB),NOSPLIT,$0-0
	MOVL	cycles+0(FP), AX
again:
	PAUSE
	SUBL	$1, AX
	JNZ	again
	RET
```

> 为什么使用 PAUSE 指令呢？
> PAUSE 指令会告诉 CPU 我当前处于处于自旋状态，这时候 CPU 会针对性的做一些优化，并且在执行这个指令的时候 CPU 会降低自己的功耗，减少能源消耗

```
if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
	atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
	awoke = true
}
```

在自旋的过程中会尝试设置 `mutexWoken` 来通知解锁，从而避免唤醒其他已经休眠的 `goroutine` 在自旋模式下，当前的 `goroutine` 就能更快的获取到锁

```go
new := old
// Don't try to acquire starving mutex, new arriving goroutines must queue.
if old&mutexStarving == 0 {
	new |= mutexLocked
}
if old&(mutexLocked|mutexStarving) != 0 {
	new += 1 << mutexWaiterShift
}
// The current goroutine switches mutex to starvation mode.
// But if the mutex is currently unlocked, don't do the switch.
// Unlock expects that starving mutex has waiters, which will not
// be true in this case.
if starving && old&mutexLocked != 0 {
	new |= mutexStarving
}
if awoke {
	// The goroutine has been woken from sleep,
	// so we need to reset the flag in either case.
	if new&mutexWoken == 0 {
		throw("sync: inconsistent mutex state")
	}
	new &^= mutexWoken
}
```

自旋结束之后就会去计算当前互斥锁的状态，如果当前处在饥饿模式下则不会去请求锁，而是会将当前 goroutine 放到队列的末端

```go
if atomic.CompareAndSwapInt32(&m.state, old, new) {
    if old&(mutexLocked|mutexStarving) == 0 {
        break // locked the mutex with CAS
    }
    // If we were already waiting before, queue at the front of the queue.
    queueLifo := waitStartTime != 0
    if waitStartTime == 0 {
        waitStartTime = runtime_nanotime()
    }
    runtime_SemacquireMutex(&m.sema, queueLifo, 1)
    starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs
    old = m.state
    if old&mutexStarving != 0 {
        // If this goroutine was woken and mutex is in starvation mode,
        // ownership was handed off to us but mutex is in somewhat
        // inconsistent state: mutexLocked is not set and we are still
        // accounted as waiter. Fix that.
        if old&(mutexLocked|mutexWoken) != 0 || old>>mutexWaiterShift == 0 {
            throw("sync: inconsistent mutex state")
        }
        delta := int32(mutexLocked - 1<<mutexWaiterShift)
        if !starving || old>>mutexWaiterShift == 1 {
            // Exit starvation mode.
            // Critical to do it here and consider wait time.
            // Starvation mode is so inefficient, that two goroutines
            // can go lock-step infinitely once they switch mutex
            // to starvation mode.
            delta -= mutexStarving
        }
        atomic.AddInt32(&m.state, delta)
        break
    }
    awoke = true
    iter = 0
}
```

状态计算完成之后就会尝试使用 CAS 操作获取锁，如果获取成功就会直接退出循环
`runtime_SemacquireMutex(&m.sema, queueLifo, 1)``runtime_SemacquireMutex`



- 不断调用尝试获取锁
- 休眠当前 goroutine
- 等待信号量，唤醒 goroutine

goroutine 被唤醒之后就会去判断当前是否处于饥饿模式，如果当前等待超过 `1ms` 就会进入饥饿模式

- 饥饿模式下：会获得互斥锁，如果等待队列中只存在当前 Goroutine，互斥锁还会从饥饿模式中退出
- 正常模式下：会设置唤醒和饥饿标记、重置迭代次数并重新执行获取锁的循环

### 解锁

和加锁比解锁就很简单了，直接看注释就好

```go
// 解锁一个没有锁定的互斥量会报运行时错误
// 解锁没有绑定关系，可以一个 goroutine 锁定，另外一个 goroutine 解锁
func (m *Mutex) Unlock() {
	// Fast path: 直接尝试设置 state 的值，进行解锁
	new := atomic.AddInt32(&m.state, -mutexLocked)
    // 如果减去了 mutexLocked 的值之后不为零就会进入慢速通道，这说明有可能失败了，或者是还有其他的 goroutine 等着
	if new != 0 {
		// Outlined slow path to allow inlining the fast path.
		// To hide unlockSlow during tracing we skip one extra frame when tracing GoUnblock.
		m.unlockSlow(new)
	}
}

func (m *Mutex) unlockSlow(new int32) {
    // 解锁一个没有锁定的互斥量会报运行时错误
	if (new+mutexLocked)&mutexLocked == 0 {
		throw("sync: unlock of unlocked mutex")
	}
    // 判断是否处于饥饿模式
	if new&mutexStarving == 0 {
        // 正常模式
		old := new
		for {
			// 如果当前没有等待者.或者 goroutine 已经被唤醒或者是处于锁定状态了，就直接返回
			if old>>mutexWaiterShift == 0 || old&(mutexLocked|mutexWoken|mutexStarving) != 0 {
				return
			}
			// 唤醒等待者并且移交锁的控制权
			new = (old - 1<<mutexWaiterShift) | mutexWoken
			if atomic.CompareAndSwapInt32(&m.state, old, new) {
				runtime_Semrelease(&m.sema, false, 1)
				return
			}
			old = m.state
		}
	} else {
		// 饥饿模式，走 handoff 流程，直接将锁交给下一个等待的 goroutine，注意这个时候不会从饥饿模式中退出
		runtime_Semrelease(&m.sema, true, 1)
	}
}
```

# RWMutex

读写锁相对于互斥锁来说粒度更细，使用读写锁可以并发读，但是不能并发读写，或者并发写写

|      | **读** | **写** |
| :--: | :----: | :----: |
|  读  |   Y    |   N    |
|  写  |   N    |   N    |

## 案例

其实大部分的业务应用都是读多写少的场景，这个时候使用读写锁的性能就会比互斥锁要好一些，例如下面的这个例子，是一个配置读写的例子，我们分别使用读写锁和互斥锁实现

```go
// RWMutexConfig 读写锁实现
type RWMutexConfig struct {
	rw   sync.RWMutex
	data []int
}

// Get get config data
func (c *RWMutexConfig) Get() []int {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.data
}

// Set set config data
func (c *RWMutexConfig) Set(n []int) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.data = n
}
```

互斥锁实现

```

// MutexConfig 互斥锁实现
type MutexConfig struct {
	data []int
	mu   sync.Mutex
}

// Get get config data
func (c *MutexConfig) Get() []int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data
}

// Set set config data
func (c *MutexConfig) Set(n []int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = n
}
```

并发基准测试

```go
type iConfig interface {
	Get() []int
	Set([]int)
}

func bench(b *testing.B, c iConfig) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			c.Set([]int{100})
			c.Get()
			c.Get()
			c.Get()
			c.Set([]int{100})
			c.Get()
			c.Get()
		}
	})
}

func BenchmarkMutexConfig(b *testing.B) {
	conf := &MutexConfig{data: []int{1, 2, 3}}
	bench(b, conf)
}

func BenchmarkRWMutexConfig(b *testing.B) {
	conf := &RWMutexConfig{data: []int{1, 2, 3}}
	bench(b, conf)
}go
```

执行结果

```
❯ go test -race -bench=.
goos: linux
goarch: amd64
pkg: github.com/mohuishou/go-training/Week03/blog/04_sync/02_rwmutex
BenchmarkMutexConfig-4            179577              6912 ns/op
BenchmarkRWMutexConfig-4          341620              3425 ns/op
PASS
ok      github.com/mohuishou/go-training/Week03/blog/04_sync/02_rwmutex 3.565s
```

可以看到首先是没有 data race 问题，其次读写锁的性能几乎是互斥锁的一倍

## 源码解析

### 基本结构

```
type RWMutex struct {
	w           Mutex  // 复用互斥锁
	writerSem   uint32 // 信号量，用于写等待读
	readerSem   uint32 // 信号量，用于读等待写
	readerCount int32  // 当前执行读的 goroutine 数量
	readerWait  int32  // 写操作被阻塞的准备读的 goroutine 的数量
}
```

由于复用了互斥锁的代码，读写锁的源码很简单，这里我就不单独画图了

### 读锁

#### 加锁

```go
func (rw *RWMutex) RLock() {
	if atomic.AddInt32(&rw.readerCount, 1) < 0 {
		// A writer is pending, wait for it.
		runtime_SemacquireMutex(&rw.readerSem, false, 0)
	}
}
```

首先是读锁， `atomic.AddInt32(&rw.readerCount, 1)` 调用这个原子方法，对当前在读的数量加一，如果返回负数，那么说明当前有其他写锁，这时候就调用 `runtime_SemacquireMutex` 休眠 goroutine 等待被唤醒

#### 解锁

```go
func (rw *RWMutex) RUnlock() {
	if r := atomic.AddInt32(&rw.readerCount, -1); r < 0 {
		// Outlined slow-path to allow the fast-path to be inlined
		rw.rUnlockSlow(r)
	}
}
```

解锁的时候对正在读的操作减一，如果返回值小于 0 那么说明当前有在写的操作，这个时候调用 `rUnlockSlow` 进入慢速通道

```go
func (rw *RWMutex) rUnlockSlow(r int32) {
	if r+1 == 0 || r+1 == -rwmutexMaxReaders {
		race.Enable()
		throw("sync: RUnlock of unlocked RWMutex")
	}
	// A writer is pending.
	if atomic.AddInt32(&rw.readerWait, -1) == 0 {
		// The last reader unblocks the writer.
		runtime_Semrelease(&rw.writerSem, false, 1)
	}
}
```

被阻塞的准备读的 goroutine 的数量减一，readerWait 为 0，就表示当前没有正在准备读的 goroutine 这时候调用 `runtime_Semrelease` 唤醒写操作

### 写锁

#### 加锁

```go
func (rw *RWMutex) Lock() {
	// First, resolve competition with other writers.
	rw.w.Lock()
	// Announce to readers there is a pending writer.
	r := atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders) + rwmutexMaxReaders
	// Wait for active readers.
	if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
		runtime_SemacquireMutex(&rw.writerSem, false, 0)
	}
}
```

首先调用互斥锁的 lock，获取到互斥锁之后，

- `atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders)` 调用这个函数阻塞后续的读操作
- 如果计算之后当前仍然有其他 goroutine 持有读锁，那么就调用 `runtime_SemacquireMutex` 休眠当前的 goroutine 等待所有的读操作完成

#### 解锁

```go
func (rw *RWMutex) Unlock() {
	// Announce to readers there is no active writer.
	r := atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)
	if r >= rwmutexMaxReaders {
		race.Enable()
		throw("sync: Unlock of unlocked RWMutex")
	}
	// Unblock blocked readers, if any.
	for i := 0; i < int(r); i++ {
		runtime_Semrelease(&rw.readerSem, false, 0)
	}
}
```

解锁的操作，会先调用 `atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)` 将恢复之前写入的负数，然后根据当前有多少个读操作在等待，循环唤醒

# 参考文献

1. https://pkg.go.dev/sync 官网文档，必读
2. https://pkg.go.dev/golang.org/x/sync@v0.0.0-20201207232520-09787c993a3a/errgroup 官网文档，必读
3. https://pkg.go.dev/sync/atomic 官网文档，必读
4. [Go: How to Reduce Lock Contention with the Atomic Package](https://medium.com/a-journey-with-go/go-how-to-reduce-lock-contention-with-the-atomic-package-ba3b2664b549)
5. [Go: Mutex and Starvation](https://medium.com/a-journey-with-go/go-mutex-and-starvation-3f4f4e75ad50)
6. [Go 进阶 27:Go 语言 Mutex Starvation(译)](https://mojotv.cn/go/golang-muteex-starvation)
7. [Go 语言设计与实现-6.2 同步原语与锁](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-sync-primitives/) 这本书值得一看
8. [PAUSE — Spin Loop Hint](https://www.felixcloutier.com/x86/pause)
9. [Linux x86 自旋锁的实现](https://github.com/freelancer-leon/notes/blob/master/kernel/lock/Lock-2-Linux_x86_Spin_Lock.md#为什么不是-pause-指令)
10. https://github.com/golang/go/commit/0556e26273f704db73df9e7c4c3d2e8434dec7be