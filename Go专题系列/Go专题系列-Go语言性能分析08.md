# 用GODEBUG看GC

## 一、GC基础知识

### 什么是 GC

在计算机科学中，垃圾回收（GC）是一种自动管理内存的机制，垃圾回收器会去尝试回收程序不再使用的对象及其占用的内存。而最早 John McCarthy 在 1959 年左右发明了垃圾回收，以简化 Lisp 中的手动内存管理的机制（来自 wikipedia）。

### 为什么要 GC

手动管理内存非常麻烦，如果管错或者管漏内存，将会直接导致程序不稳定（持续泄露），甚至立刻崩溃。

### GC 带来的问题

硬要说会带来什么问题的话，也就数大家最关注的 Stop The World（STW），STW 代指在执行某个垃圾回收算法的某个阶段时，需要将整个应用程序暂停去处理 GC 相关的工作事项。例如：

| 行为       | 会不会 STW | 为什么                                                       |
| :--------- | :--------- | :----------------------------------------------------------- |
| 标记开始   | 会         | 在开始标记时，准备根对象的扫描，会打开写屏障（Write Barrier） 和 辅助GC（mutator assist），而回收器和应用程序是并发运行的，因此会暂停当前正在运行的所有 Goroutine。 |
| 并发标记中 | 不会       | 标记阶段，主要目的是标记堆内存中仍在使用的值。               |
| 标记结束   | 会         | 在完成标记任务后，将重新扫描部分根对象，这时候会禁用写屏障（Write Barrier）和辅助GC（mutator assist），而标记阶段和应用程序是并发运行的，所以在标记阶段可能会有新的对象产生，因此在重新扫描时需要进行 STW。 |

### 如何调整 GC 频率

可以通过 GOGC 变量设置初始垃圾收集器的目标百分比值，对比的规则为当新分配的数值与上一次收集后剩余的实时数值的比例达到设置的目标百分比时，就会触发 GC，默认值为 GOGC=100。如果将其设置为 GOGC=off 可以完全禁用垃圾回收器，要不试试？

简单来讲就是，GOGC 的值设置的越大，GC 的频率越低，但每次最终所触发到 GC 的堆内存也会更大。

在程序运行时，可以通过调用下述方法来动态调整GOGC的值：

```
// runtime/debug
debug.SetGCPercent
```



### 各版本 GC 情况

| 版本    | GC 算法                                                      | STW 时间               |
| :------ | :----------------------------------------------------------- | :--------------------- |
| Go 1.0  | STW（强依赖 tcmalloc）                                       | 百ms到秒级别           |
| Go 1.3  | Mark STW, Sweep 并行                                         | 百ms级别               |
| Go 1.5  | 三色标记法, 并发标记清除。同时运行时从 C 和少量汇编，改为 Go 和少量汇编实现 | 10-50ms级别            |
| Go 1.6  | 1.5 中一些与并发 GC 不协调的地方更改，集中式的 GC 协调协程，改为状态机实现 | 5ms级别                |
| Go 1.7  | GC 时由 mark 栈收缩改为并发，span 对象分配机制由 freelist 改为 bitmap 模式，SSA引入 | ms级别                 |
| Go 1.8  | 混合写屏障（hybrid write barrier）, 消除 re-scanning stack   | sub ms                 |
| Go 1.12 | Mark Termination 流程优化                                    | sub ms, 但几乎减少一半 |

注：资料来源于 @boya 在深圳 Gopher Meetup 的分享。

### 主动触发 GC

在 Go 语言中，与 GC 相关的 API 极少，如果想要主动触发 GC 行为，则可以通过调用`runtime.GC `方法来达到这个目的。与之相配套的内存相关的归还，也可以通过手动调用`debug.FreeOSMemory`方法来实现。

## 二、GODEBUG

GODEBUG 变量可以控制运行时内的调试变量，参数以逗号分隔，格式为：`name=val`。本文着重点在 GC 的观察上，主要涉及 gctrace 参数，我们通过设置 `gctrace=1` 后就可以使得垃圾收集器向标准错误流发出 GC 运行信息。

### 涉及术语

- mark：标记阶段。
- markTermination：标记结束阶段。
- mutator assist：辅助 GC，是指在 GC 过程中 mutator 线程会并发运行，而 mutator assist 机制会协助 GC 做一部分的工作。
- heaplive：在 Go 的内存管理中，span 是内存页的基本单元，每页大小为 8kb，同时 Go 会根据对象的大小不同而分配不同页数的 span，而 heaplive 就代表着所有 span 的总大小。
- dedicated / fractional / idle：在标记阶段会分为三种不同的 mark worker 模式，分别是 dedicated、fractional 和 idle，它们代表着不同的专注程度，其中 dedicated 模式最专注，是完整的 GC 回收行为，fractional 只会干部分的 GC 行为，idle 最轻松。这里你只需要了解它是不同专注程度的 mark worker 就好了，详细介绍我们可以等后续的文章。

### 示例代码

创建一个demo.go文件，写入如下代码：

```javascript
package main

import "sync"

func main() {
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(wg *sync.WaitGroup) {
			var counter int
			for i := 0; i < 1e10; i++ {
				counter++
			}
			wg.Done()
		}(&wg)
	}
	wg.Wait()
}
```

### gctrace

在命令行设置 `GODEBUG=gctrace=1` 来运行 demo.go 文件，为了便于展示和说明，这里把gctrace的输出结果拆成两部分，分别是GC信息和Scavenging信息。

#### （1）GC信息：

```javascript
$ GODEBUG=gctrace=1 go run demo1.go
gc 1 @0.361s 3%: 0+28+0 ms clock, 0+28/0/28+0 ms cpu, 4->4->0 MB, 5 MB goal, 2 P
gc 2 @0.587s 2%: 0+3.0+0 ms clock, 0+0/0/3.0+0 ms cpu, 4->4->0 MB, 5 MB goal, 2 P
gc 3 @0.792s 1%: 0+1.0+0 ms clock, 0+1.0/0/1.0+0 ms cpu, 4->4->0 MB, 5 MB goal, 2 P
scvg: 0 MB released
scvg: inuse: 1, idle: 6, sys: 7, released: 3, consumed: 4 (MB)
scvg: inuse: 1, idle: 6, sys: 7, released: 3, consumed: 4 (MB)
scvg: 0 MB released
............
```

### 格式

```javascript
gc # @#s #%: #+#+# ms clock, #+#/#/#+# ms cpu, #->#-># MB, # MB goal, # P
```

### 含义

- `gc#`：GC 执行次数的编号，每次叠加。
- `@#s`：自程序启动后到当前的具体秒数。
- `#%`：自程序启动以来在GC中花费的时间百分比。
- `#+...+#`：GC 的标记工作共使用的 CPU 时间占总 CPU 时间的百分比。
- `#->#-># MB`：分别表示 GC 启动时, GC 结束时, GC 活动时的堆大小.
- `#MB goal`：下一次触发 GC 的内存占用阈值。
- `#P`：当前使用的处理器 P 的数量。

### 案例

```javascript
gc 7 @0.140s 1%: 0.031+2.0+0.042 ms clock, 0.12+0.43/1.8/0.049+0.17 ms cpu, 4->4->1 MB, 5 MB goal, 4 P
```

- gc 7：第 7 次 GC。
- @0.140s：当前是程序启动后的 0.140s。
- 1%：程序启动后到现在共花费 1% 的时间在 GC 上。
- 0.031+2.0+0.042 ms clock：
- 0.031：表示单个 P 在 mark 阶段的 STW 时间。
- 2.0：表示所有 P 的 mark concurrent（并发标记）所使用的时间。
- 0.042：表示单个 P 的 markTermination 阶段的 STW 时间。
- 0.12+0.43/1.8/0.049+0.17 ms cpu：
- 0.12：表示整个进程在 mark 阶段 STW 停顿的时间。
- 0.43/1.8/0.049：0.43 表示 mutator assist 占用的时间，1.8 表示 dedicated + fractional 占用的时间，0.049 表示 idle 占用的时间。
- 0.17ms：0.17 表示整个进程在 markTermination 阶段 STW 时间。
- 4->4->1 MB：
- 4：表示开始 mark 阶段前的 heap_live 大小。
- 4：表示开始 markTermination 阶段前的 heap_live 大小。
- 1：表示被标记对象的大小。
- 5 MB goal：表示下一次触发 GC 回收的阈值是 5 MB。
- 4 P：本次 GC 一共涉及多少个 P。

## 总结

通过本章节我们掌握了使用 GODEBUG 查看应用程序 GC 运行情况的方法，只要用这种方法我们就可以观测不同情况下 GC 的情况了，甚至可以做出非常直观的对比图，大家不妨尝试一下。