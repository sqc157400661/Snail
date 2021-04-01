## 曹春晖：谈一谈 Go 和 Syscall

> **导读：**syscall 是语言与系统交互的唯一手段，理解 Go 语言中的 syscall，本文可以帮助读者理解 Go 语言怎么与系统打交道，同时了解底层 runtime 在 syscall 优化方面的一些小心思，从而更为深入地理解 Go 语言。


  

## **阅读索引**



- 概念

- 入口

- 系统调用管理

- runtime 中的 SYSCALL

- 和调度的交互

- - entersyscall
  - exitsyscallfast
  - exitsyscall
  - entersyscallblock
  - entersyscallblock_handoff
  - entersyscall_sysmon
  - entersyscall_gcwait

- 总结

## **概念**

![图片](D:\www\Snail\Go专题系列\book\images\498456sdfsadfsd.png)

## **入口**

syscall 有下面几个入口，在 syscall/asm_linux_amd64.s 中。



```
func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)
func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
func RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)
```



这些函数的实现都是汇编，按照 linux 的 syscall 调用规范，我们只要在汇编中把参数依次传入寄存器，并调用 SYSCALL 指令即可进入内核处理逻辑，系统调用执行完毕之后，返回值放在 RAX 中：



![图片](D:\www\Snail\Go专题系列\book\images\54645dfsaffghfgjhgjgh.png)



Syscall 和 Syscall6 的区别只有传入参数不一样:



```
1// func Syscall(trap int64, a1, a2, a3 uintptr) (r1, r2, err uintptr);
 2TEXT ·Syscall(SB),NOSPLIT,$0-56
 3    CALL    runtime·entersyscall(SB)
 4    MOVQ    a1+8(FP), DI
 5    MOVQ    a2+16(FP), SI
 6    MOVQ    a3+24(FP), DX
 7    MOVQ    $0, R10
 8    MOVQ    $0, R8
 9    MOVQ    $0, R9
10    MOVQ    trap+0(FP), AX    // syscall entry
11    SYSCALL
12    // 0xfffffffffffff001 是 linux MAX_ERRNO 取反 转无符号，http://lxr.free-electrons.com/source/include/linux/err.h#L17
13    CMPQ    AX, $0xfffffffffffff001
14    JLS    ok
15    MOVQ    $-1, r1+32(FP)
16    MOVQ    $0, r2+40(FP)
17    NEGQ    AX
18    MOVQ    AX, err+48(FP)
19    CALL    runtime·exitsyscall(SB)
20    RET
21ok:
22    MOVQ    AX, r1+32(FP)
23    MOVQ    DX, r2+40(FP)
24    MOVQ    $0, err+48(FP)
25    CALL    runtime·exitsyscall(SB)
26    RET
27
28// func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr)
29TEXT ·Syscall6(SB),NOSPLIT,$0-80
30    CALL    runtime·entersyscall(SB)
31    MOVQ    a1+8(FP), DI
32    MOVQ    a2+16(FP), SI
33    MOVQ    a3+24(FP), DX
34    MOVQ    a4+32(FP), R10
35    MOVQ    a5+40(FP), R8
36    MOVQ    a6+48(FP), R9
37    MOVQ    trap+0(FP), AX    // syscall entry
38    SYSCALL
39    CMPQ    AX, $0xfffffffffffff001
40    JLS    ok6
41    MOVQ    $-1, r1+56(FP)
42    MOVQ    $0, r2+64(FP)
43    NEGQ    AX
44    MOVQ    AX, err+72(FP)
45    CALL    runtime·exitsyscall(SB)
46    RET
47ok6:
48    MOVQ    AX, r1+56(FP)
49    MOVQ    DX, r2+64(FP)
50    MOVQ    $0, err+72(FP)
51    CALL    runtime·exitsyscall(SB)
52    RET
```



两个函数没什么大区别，为啥不用一个呢？个人猜测，Go 的函数参数都是栈上传入，可能是为了节省一点栈空间。。在正常的 Syscall 操作之前会通知 runtime，接下来我要进行 syscall 操作了 **runtime·entersyscall** ，退出时会调用 **runtime·exitsyscall** 。



```
1// func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2, err uintptr)
 2TEXT ·RawSyscall(SB),NOSPLIT,$0-56
 3    MOVQ    a1+8(FP), DI
 4    MOVQ    a2+16(FP), SI
 5    MOVQ    a3+24(FP), DX
 6    MOVQ    $0, R10
 7    MOVQ    $0, R8
 8    MOVQ    $0, R9
 9    MOVQ    trap+0(FP), AX    // syscall entry
10    SYSCALL
11    CMPQ    AX, $0xfffffffffffff001
12    JLS    ok1
13    MOVQ    $-1, r1+32(FP)
14    MOVQ    $0, r2+40(FP)
15    NEGQ    AX
16    MOVQ    AX, err+48(FP)
17    RET
18ok1:
19    MOVQ    AX, r1+32(FP)
20    MOVQ    DX, r2+40(FP)
21    MOVQ    $0, err+48(FP)
22    RET
23
24// func RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr)
25TEXT ·RawSyscall6(SB),NOSPLIT,$0-80
26    MOVQ    a1+8(FP), DI
27    MOVQ    a2+16(FP), SI
28    MOVQ    a3+24(FP), DX
29    MOVQ    a4+32(FP), R10
30    MOVQ    a5+40(FP), R8
31    MOVQ    a6+48(FP), R9
32    MOVQ    trap+0(FP), AX    // syscall entry
33    SYSCALL
34    CMPQ    AX, $0xfffffffffffff001
35    JLS    ok2
36    MOVQ    $-1, r1+56(FP)
37    MOVQ    $0, r2+64(FP)
38    NEGQ    AX
39    MOVQ    AX, err+72(FP)
40    RET
41ok2:
42    MOVQ    AX, r1+56(FP)
43    MOVQ    DX, r2+64(FP)
44    MOVQ    $0, err+72(FP)
45    RET
```



RawSyscall 和 Syscall 的区别也非常微小，就只是在进入 Syscall 和退出的时候没有通知 runtime，这样 runtime 理论上是没有办法通过调度把这个 g 的 m 的 p 调度走的，所以如果用户代码使用了 RawSyscall 来做一些阻塞的系统调用，是有可能阻塞其它的 g 的，下面是官方开发的原话:



*Yes, if you call RawSyscall you may block other goroutines from running. The system monitor may start them up after a while, but I think there are cases where it won't. I would say that Go programs should always call Syscall. RawSyscall exists to make it slightly more efficient to call system calls that never block, such as getpid. But it's really an internal mechanism.*



```
1// func gettimeofday(tv *Timeval) (err uintptr)
 2TEXT ·gettimeofday(SB),NOSPLIT,$0-16
 3    MOVQ    tv+0(FP), DI
 4    MOVQ    $0, SI
 5    MOVQ    runtime·__vdso_gettimeofday_sym(SB), AX
 6    CALL    AX
 7
 8    CMPQ    AX, $0xfffffffffffff001
 9    JLS    ok7
10    NEGQ    AX
11    MOVQ    AX, err+8(FP)
12    RET
13ok7:
14    MOVQ    $0, err+8(FP)
15    RET
```





## ▎**系统调用管理**



先是系统调用的定义文件:

```
/syscall/syscall_linux.go
```



可以把系统调用分为三类:



- 阻塞系统调用
- 非阻塞系统调用
- wrapped 系统调用



阻塞系统调用会定义成下面这样的形式:



```
//sys   Madvise(b []byte, advice int) (err error)
```



非阻塞系统调用:



```
//sysnb    EpollCreate(size int) (fd int, err error)
```



然后，根据这些注释，mksyscall.pl 脚本会生成对应的平台的具体实现。mksyscall.pl 是一段 perl 脚本，感兴趣的同学可以自行查看，这里就不再赘述了。



看看阻塞和非阻塞的系统调用的生成结果:



```
 1func Madvise(b []byte, advice int) (err error) {
 2    var _p0 unsafe.Pointer
 3    if len(b) > 0 {
 4        _p0 = unsafe.Pointer(&b[0])
 5    } else {
 6        _p0 = unsafe.Pointer(&_zero)
 7    }
 8    _, _, e1 := Syscall(SYS_MADVISE, uintptr(_p0), uintptr(len(b)), uintptr(advice))
 9    if e1 != 0 {
10        err = errnoErr(e1)
11    }
12    return
13}
14
15func EpollCreate(size int) (fd int, err error) {
16    r0, _, e1 := RawSyscall(SYS_EPOLL_CREATE, uintptr(size), 0, 0)
17    fd = int(r0)
18    if e1 != 0 {
19        err = errnoErr(e1)
20    }
21    return
22}
```



显然，标记为 sys 的系统调用使用的是 Syscall 或者 Syscall6，标记为 sysnb 的系统调用使用的是 RawSyscall 或 RawSyscall6。

wrapped 的系统调用是怎么一回事呢？

```
1func Rename(oldpath string, newpath string) (err error) {
2    return Renameat(_AT_FDCWD, oldpath, _AT_FDCWD, newpath)
3}
```



可能是觉得系统调用的名字不太好，或者参数太多，我们就简单包装一下。没啥特别的。





## ▎**runtime 中的 SYSCALL**



除了上面提到的阻塞非阻塞和 wrapped syscall，runtime 中还定义了一些 low-level 的 syscall，这些是不暴露给用户的。



提供给用户的 syscall 库，在使用时，会使 goroutine 和 p 分别进入 Gsyscall 和 Psyscall 状态。但 runtime 自己封装的这些 syscall 无论是否阻塞，都不会调用 entersyscall 和 exitsyscall。虽说是 “low-level” 的 syscall。



不过和暴露给用户的 syscall 本质是一样的。这些代码在 runtime/sys_linux_amd64.s中，举个具体的例子:



```
1     TEXT runtime·write(SB),NOSPLIT,$0-28
 2    MOVQ    fd+0(FP), DI
 3    MOVQ    p+8(FP), SI
 4    MOVL    n+16(FP), DX
 5    MOVL    $SYS_write, AX
 6    SYSCALL
 7    CMPQ    AX, $0xfffffffffffff001
 8    JLS    2(PC)
 9    MOVL    $-1, AX
10    MOVL    AX, ret+24(FP)
11    RET
12
13TEXT runtime·read(SB),NOSPLIT,$0-28
14    MOVL    fd+0(FP), DI
15    MOVQ    p+8(FP), SI
16    MOVL    n+16(FP), DX
17    MOVL    $SYS_read, AX
18    SYSCALL
19    CMPQ    AX, $0xfffffffffffff001
20    JLS    2(PC)
21    MOVL    $-1, AX
22    MOVL    AX, ret+24(FP)
23    RET
```



下面是所有 runtime 另外定义的 syscall 列表:



```
 1#define SYS_read        0
 2#define SYS_write        1
 3#define SYS_open        2
 4#define SYS_close        3
 5#define SYS_mmap        9
 6#define SYS_munmap        11
 7#define SYS_brk         12
 8#define SYS_rt_sigaction    13
 9#define SYS_rt_sigprocmask    14
10#define SYS_rt_sigreturn    15
11#define SYS_access        21
12#define SYS_sched_yield     24
13#define SYS_mincore        27
14#define SYS_madvise        28
15#define SYS_setittimer        38
16#define SYS_getpid        39
17#define SYS_socket        41
18#define SYS_connect        42
19#define SYS_clone        56
20#define SYS_exit        60
21#define SYS_kill        62
22#define SYS_fcntl        72
23#define SYS_getrlimit        97
24#define SYS_sigaltstack     131
25#define SYS_arch_prctl        158
26#define SYS_gettid        186
27#define SYS_tkill        200
28#define SYS_futex        202
29#define SYS_sched_getaffinity    204
30#define SYS_epoll_create    213
31#define SYS_exit_group        231
32#define SYS_epoll_wait        232
33#define SYS_epoll_ctl        233
34#define SYS_pselect6        270
35#define SYS_epoll_create1    291
```



这些 syscall 理论上都是不会在执行期间被调度器剥离掉 p 的，所以执行成功之后 goroutine 会继续执行，而不像用户的 goroutine 一样，若被剥离 p 会进入等待队列。



## ▎**和调度的交互**



既然要和调度交互，那友好地通知我要 syscall 了: entersyscall，我完事了: exitsyscall。



所以这里的交互指的是用户代码使用 syscall 库时和调度器的交互。**runtime 里的 syscall 不走这套流程。**



## **▎entersyscall**



```
1// syscall 库和 cgo 调用的标准入口
 2//go:nosplit
 3func entersyscall() {
 4    reentersyscall(getcallerpc(), getcallersp())
 5}
 6
 7//go:nosplit
 8func reentersyscall(pc, sp uintptr) {
 9    _g_ := getg()
10
11    // 需要禁止 g 的抢占
12    _g_.m.locks++
13
14    // entersyscall 中不能调用任何会导致栈增长/分裂的函数
15    _g_.stackguard0 = stackPreempt
16    // 设置 throwsplit，在 newstack 中，如果发现 throwsplit 是 true
17    // 会直接 crash
18    // 下面的代码是 newstack 里的
19    // if thisg.m.curg.throwsplit {
20    //     throw("runtime: stack split at bad time")
21    // }
22    _g_.throwsplit = true
23
24    // Leave SP around for GC and traceback.
25    // 保存现场，在 syscall 之后会依据这些数据恢复现场
26    save(pc, sp)
27    _g_.syscallsp = sp
28    _g_.syscallpc = pc
29    casgstatus(_g_, _Grunning, _Gsyscall)
30    if _g_.syscallsp < _g_.stack.lo || _g_.stack.hi < _g_.syscallsp {
31        systemstack(func() {
32            print("entersyscall inconsistent ", hex(_g_.syscallsp), " [", hex(_g_.stack.lo), ",", hex(_g_.stack.hi), "]\n")
33            throw("entersyscall")
34        })
35    }
36
37    if atomic.Load(&sched.sysmonwait) != 0 {
38        systemstack(entersyscall_sysmon)
39        save(pc, sp)
40    }
41
42    if _g_.m.p.ptr().runSafePointFn != 0 {
43        // runSafePointFn may stack split if run on this stack
44        systemstack(runSafePointFn)
45        save(pc, sp)
46    }
47
48    _g_.m.syscalltick = _g_.m.p.ptr().syscalltick
49    _g_.sysblocktraced = true
50    _g_.m.mcache = nil
51    _g_.m.p.ptr().m = 0
52    atomic.Store(&_g_.m.p.ptr().status, _Psyscall)
53    if sched.gcwaiting != 0 {
54        systemstack(entersyscall_gcwait)
55        save(pc, sp)
56    }
57
58    _g_.m.locks--
59}
```



可以看到，进入 syscall 的 G 是铁定不会被抢占的。



## **▎****exitsyscall**



```
1// g 已经退出了 syscall
 2// 需要准备让 g 在 cpu 上重新运行
 3// 这个函数只会在 syscall 库中被调用，在 runtime 里用的 low-level syscall
 4// 不会用到
 5// 不能有 write barrier，因为 P 可能已经被偷走了
 6//go:nosplit
 7//go:nowritebarrierrec
 8func exitsyscall(dummy int32) {
 9    _g_ := getg()
10
11    _g_.m.locks++ // see comment in entersyscall
12    if getcallersp(unsafe.Pointer(&dummy)) > _g_.syscallsp {
13        // throw calls print which may try to grow the stack,
14        // but throwsplit == true so the stack can not be grown;
15        // use systemstack to avoid that possible problem.
16        systemstack(func() {
17            throw("exitsyscall: syscall frame is no longer valid")
18        })
19    }
20
21    _g_.waitsince = 0
22    oldp := _g_.m.p.ptr()
23    if exitsyscallfast() {
24        if _g_.m.mcache == nil {
25            systemstack(func() {
26                throw("lost mcache")
27            })
28        }
29        // 目前有 p，可以运行
30        _g_.m.p.ptr().syscalltick++
31        // 把 g 的状态修改回 running
32        casgstatus(_g_, _Gsyscall, _Grunning)
33
34        // 垃圾收集未在运行(因为我们这段逻辑在执行)
35        // 所以清理掉 syscallsp 是安全的
36        _g_.syscallsp = 0
37        _g_.m.locks--
38        if _g_.preempt {
39            // 防止在 newstack 中清理掉 preemption 标记
40            _g_.stackguard0 = stackPreempt
41        } else {
42            // 否则恢复在 entersyscall/entersyscallblock 中破坏掉的正常的 _StackGuard
43            _g_.stackguard0 = _g_.stack.lo + _StackGuard
44        }
45        _g_.throwsplit = false
46        return
47    }
48
49    _g_.sysexitticks = 0
50    _g_.m.locks--
51
52    // 调用 scheduler
53    mcall(exitsyscall0)
54
55    if _g_.m.mcache == nil {
56        systemstack(func() {
57            throw("lost mcache")
58        })
59    }
60
61    // 调度器返回了，所以我们可以清理掉在 syscall 期间为垃圾收集器
62    // 准备的 syscallsp 信息了
63    // 需要一直等待到 gosched 返回，我们不确定垃圾收集器是不是在运行
64    _g_.syscallsp = 0
65    _g_.m.p.ptr().syscalltick++
66    _g_.throwsplit = false
67}
```



这里还调用了 exitsyscallfast 和 exitsyscall0。



## **▎****exitsyscallfast**



```
1//go:nosplit
 2func exitsyscallfast() bool {
 3    _g_ := getg()
 4
 5    // Freezetheworld sets stopwait but does not retake P's.
 6    if sched.stopwait == freezeStopWait {
 7        _g_.m.mcache = nil
 8        _g_.m.p = 0
 9        return false
10    }
11
12    // Try to re-acquire the last P.
13    if _g_.m.p != 0 && _g_.m.p.ptr().status == _Psyscall && atomic.Cas(&_g_.m.p.ptr().status, _Psyscall, _Prunning) {
14        // There's a cpu for us, so we can run.
15        exitsyscallfast_reacquired()
16        return true
17    }
18
19    // Try to get any other idle P.
20    oldp := _g_.m.p.ptr()
21    _g_.m.mcache = nil
22    _g_.m.p = 0
23    if sched.pidle != 0 {
24        var ok bool
25        systemstack(func() {
26            ok = exitsyscallfast_pidle()
27        })
28        if ok {
29            return true
30        }
31    }
32    return false
33}
```



总之就是努力获取一个 P 来执行 syscall 之后的逻辑。如果哪都没有 P 可以给我们用，那就进入 exitsyscall0 了。



```
1mcall(exitsyscall0)
```



调用 exitsyscall0 时，会切换到 g0 栈。



## **▎****exitsyscall0**



```
1// 在 exitsyscallfast 中吃瘪了，没办法，慢慢来
 2// 把 g 的状态设置成 runnable，先进 runq 等着
 3//go:nowritebarrierrec
 4func exitsyscall0(gp *g) {
 5    _g_ := getg()
 6
 7    casgstatus(gp, _Gsyscall, _Grunnable)
 8    dropg()
 9    lock(&sched.lock)
10    _p_ := pidleget()
11    if _p_ == nil {
12        // 如果 P 被人偷跑了
13        globrunqput(gp)
14    } else if atomic.Load(&sched.sysmonwait) != 0 {
15        atomic.Store(&sched.sysmonwait, 0)
16        notewakeup(&sched.sysmonnote)
17    }
18    unlock(&sched.lock)
19    if _p_ != nil {
20        // 如果现在还有 p，那就用这个 p 执行
21        acquirep(_p_)
22        execute(gp, false) // Never returns.
23    }
24    if _g_.m.lockedg != 0 {
25        // 设置了 LockOsThread 的 g 的特殊逻辑
26        stoplockedm()
27        execute(gp, false) // Never returns.
28    }
29    stopm()
30    schedule() // Never returns.
31}
```



## **▎****entersyscallblock**



知道自己会 block，直接就把 p 交出来了。



```
 1// 和 entersyscall 一样，就是会直接把 P 给交出去，因为知道自己是会阻塞的
 2//go:nosplit
 3func entersyscallblock(dummy int32) {
 4    _g_ := getg()
 5
 6    _g_.m.locks++ // see comment in entersyscall
 7    _g_.throwsplit = true
 8    _g_.stackguard0 = stackPreempt // see comment in entersyscall
 9    _g_.m.syscalltick = _g_.m.p.ptr().syscalltick
10    _g_.sysblocktraced = true
11    _g_.m.p.ptr().syscalltick++
12
13    // Leave SP around for GC and traceback.
14    pc := getcallerpc()
15    sp := getcallersp(unsafe.Pointer(&dummy))
16    save(pc, sp)
17    _g_.syscallsp = _g_.sched.sp
18    _g_.syscallpc = _g_.sched.pc
19    if _g_.syscallsp < _g_.stack.lo || _g_.stack.hi < _g_.syscallsp {
20        sp1 := sp
21        sp2 := _g_.sched.sp
22        sp3 := _g_.syscallsp
23        systemstack(func() {
24            print("entersyscallblock inconsistent ", hex(sp1), " ", hex(sp2), " ", hex(sp3), " [", hex(_g_.stack.lo), ",", hex(_g_.stack.hi), "]\n")
25            throw("entersyscallblock")
26        })
27    }
28    casgstatus(_g_, _Grunning, _Gsyscall)
29    if _g_.syscallsp < _g_.stack.lo || _g_.stack.hi < _g_.syscallsp {
30        systemstack(func() {
31            print("entersyscallblock inconsistent ", hex(sp), " ", hex(_g_.sched.sp), " ", hex(_g_.syscallsp), " [", hex(_g_.stack.lo), ",", hex(_g_.stack.hi), "]\n")
32            throw("entersyscallblock")
33        })
34    }
35
36    // 直接调用 entersyscallblock_handoff 把 p 交出来了
37    systemstack(entersyscallblock_handoff)
38
39    // Resave for traceback during blocked call.
40    save(getcallerpc(), getcallersp(unsafe.Pointer(&dummy)))
41
42    _g_.m.locks--
43}
```



这个函数只有一个调用方 notesleepg，这里就不再赘述了。



## **▎entersyscallblock_handoff**



```
1func entersyscallblock_handoff() {
2    handoffp(releasep())
3}
```



比较简单。



## **▎entersyscall_sysmon**



```
1func entersyscall_sysmon() {
2    lock(&sched.lock)
3    if atomic.Load(&sched.sysmonwait) != 0 {
4        atomic.Store(&sched.sysmonwait, 0)
5        notewakeup(&sched.sysmonnote)
6    }
7    unlock(&sched.lock)
8}
```



## **▎****entersyscall_gcwait**



```
 1func entersyscall_gcwait() {
 2    _g_ := getg()
 3    _p_ := _g_.m.p.ptr()
 4
 5    lock(&sched.lock)
 6    if sched.stopwait > 0 && atomic.Cas(&_p_.status, _Psyscall, _Pgcstop) {
 7        _p_.syscalltick++
 8        if sched.stopwait--; sched.stopwait == 0 {
 9            notewakeup(&sched.stopnote)
10        }
11    }
12    unlock(&sched.lock)
13}
```





## **▎****总结**



提供给用户使用的系统调用，基本都会通知 runtime，以 entersyscall，exitsyscall 的形式来告诉 runtime，在这个 syscall 阻塞的时候，由 runtime 判断是否把 P 腾出来给其它的 M 用。解绑定指的是把 M 和 P 之间解绑，如果绑定被解除，在 syscall 返回时，这个 g 会被放入执行队列 runq 中。



同时 runtime 又保留了自己的特权，在执行自己的逻辑的时候，我的 P 不会被调走，这样保证了在 Go 自己“底层”使用的这些 syscall 返回之后都能被立刻处理。



所以同样是 epollwait，runtime 用的是不能被别人打断的，你用的 syscall.EpollWait 那显然是没有这种特权的。



## **▎****END**