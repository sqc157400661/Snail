# Go并发编程(八) 深入理解 sync.Once

在上一篇文章《[Week03: Go 并发编程(七) 深入理解 errgroup](https://lailin.xyz/post/go-training-week3-errgroup.html)》当中看 `errgourp` 源码的时候我们发现最后返回 `err` 是通过 once 来只保证返回一个非 nil 的值的，本文就来看一下 Once 的使用与实现

## 案例

once 的使用很简单

```go
func main() {
	var (
		o  sync.Once
		wg sync.WaitGroup
	)

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			o.Do(func() {
				fmt.Println("once", i)
			})
		}(i)
	}

	wg.Wait()
}
```

输出

```go
❯ go run ./main.go
once 9
```

## 源码分析

```go
type Once struct {
	done uint32
	m    Mutex
}
```

done 用于判定函数是否执行，如果不为 0 会直接返回

```go
func (o *Once) Do(f func()) {
	// Note: Here is an incorrect implementation of Do:
	//
	//	if atomic.CompareAndSwapUint32(&o.done, 0, 1) {
	//		f()
	//	}
	//
	// Do guarantees that when it returns, f has finished.
	// This implementation would not implement that guarantee:
	// given two simultaneous calls, the winner of the cas would
	// call f, and the second would return immediately, without
	// waiting for the first's call to f to complete.
	// This is why the slow path falls back to a mutex, and why
	// the atomic.StoreUint32 must be delayed until after f returns.

	if atomic.LoadUint32(&o.done) == 0 {
		// Outlined slow-path to allow inlining of the fast-path.
		o.doSlow(f)
	}
}
```

看 go 的源码真的可以学到很多东西，在这里还给出了很容易犯错的一种实现

```go
if atomic.CompareAndSwapUint32(&o.done, 0, 1) {
	f()
}
```

如果这么实现最大的问题是，如果并发调用，一个 `goroutine` 执行，另外一个不会等正在执行的这个成功之后返回，而是直接就返回了，**这就不能保证传入的方法一定会先执行一次了**
所以回头看官方的实现

```
if atomic.LoadUint32(&o.done) == 0 {
    // Outlined slow-path to allow inlining of the fast-path.
    o.doSlow(f)
}
```

会先判断 done 是否为 0，如果不为 0 说明还没执行过，就进入 `doSlow`

```go
func (o *Once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done == 0 {
		defer atomic.StoreUint32(&o.done, 1)
		f()
	}
}
```

在 `doSlow` 当中使用了互斥锁来保证只会执行一次

## 总结

- Once 保证了传入的函数只会执行一次，这常用在单例模式，配置文件加载，初始化这些场景下
- 但是需要注意。Once 是不能复用的，只要执行过了，再传入其他的方法也不会再执行了
- 并且 Once.Do 在执行的过程中如果 f 出现 panic，后面也不会再执行了

## 参考文献

1. https://pkg.go.dev/sync#Once
2. [6.2 同步原语与锁](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-sync-primitives/#once)