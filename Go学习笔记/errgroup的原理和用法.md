## 函数签名

```go
type Group
    func WithContext(ctx context.Context) (*Group, context.Context)
    func (g *Group) Go(f func() error)
    func (g *Group) Wait() error
```

整个包就一个 Group 结构体

- 通过 `WithContext` 可以创建一个带取消的 `Group`
- 当然除此之外也可以零值的 Group 也可以直接使用，但是出错之后就不会取消其他的 goroutine 了
- `Go` 方法传入一个 `func() error` 内部会启动一个 goroutine 去处理
- `Wait` 类似 WaitGroup 的 Wait 方法，等待所有的 goroutine 结束后退出，返回的错误是一个出错的 err

## 源码

### Group

```go
type Group struct {
    // context 的 cancel 方法
	cancel func()

    // 复用 WaitGroup
	wg sync.WaitGroup

	// 用来保证只会接受一次错误
	errOnce sync.Once
    // 保存第一个返回的错误
	err     error
}
```

### WithContext

```go
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}
```

`WithContext` 就是使用 `WithCancel` 创建一个可以取消的 context 将 cancel 赋值给 Group 保存起来，然后再将 context 返回回去

注意这里有一个坑，在后面的代码中不要把这个 ctx 当做父 context 又传给下游，因为 errgroup 取消了，这个 context 就没用了，会导致下游复用的时候出错

### Go

```go
func (g *Group) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}
```

`Go` 方法其实就类似于 `go` 关键字，会启动一个携程，然后利用 `waitgroup` 来控制是否结束，如果有一个非 `nil` 的 error 出现就会保存起来并且如果有 `cancel` 就会调用 `cancel` 取消掉，使 `ctx` 返回

### Wait

```go
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}
```

`Wait` 方法其实就是调用 `WaitGroup` 等待，如果有 `cancel` 就调用一下

## 案例

基于 errgroup 实现一个 http server 的启动和关闭 ，以及 linux signal 信号的注册和处理，要保证能够 一个退出，全部注销退出。

```go
func main() {
	g, ctx := errgroup.WithContext(context.Background())

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// 模拟单个服务错误退出
	serverOut := make(chan struct{})
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		serverOut <- struct{}{}
	})

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	// g1
	// g1 退出了所有的协程都能退出么？
	// g1 退出后, context 将不再阻塞，g2, g3 都会随之退出
	// 然后 main 函数中的 g.Wait() 退出，所有协程都会退出
	g.Go(func() error {
		return server.ListenAndServe()
	})

	// g2
	// g2 退出了所有的协程都能退出么？
	// g2 退出时，调用了 shutdown，g1 会退出
	// g2 退出后, context 将不再阻塞，g3 会随之退出
	// 然后 main 函数中的 g.Wait() 退出，所有协程都会退出
	g.Go(func() error {
		select {
		case <-ctx.Done():
			log.Println("errgroup exit...")
		case <-serverOut:
			log.Println("server will out...")
		}

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		// 这里不是必须的，但是如果使用 _ 的话静态扫描工具会报错，加上也无伤大雅
		defer cancel()

		log.Println("shutting down server...")
		return server.Shutdown(timeoutCtx)
	})

	// g3
	// g3 捕获到 os 退出信号将会退出
	// g3 退出了所有的协程都能退出么？
	// g3 退出后, context 将不再阻塞，g2 会随之退出
	// g2 退出时，调用了 shutdown，g1 会退出
	// 然后 main 函数中的 g.Wait() 退出，所有协程都会退出
	g.Go(func() error {
		quit := make(chan os.Signal, 0)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-quit:
			return errors.Errorf("get os signal: %v", sig)
		}
	})

	fmt.Printf("errgroup exiting: %+v\n", g.Wait())
}
```

这里主要用到了 errgroup 一个出错，其余取消的能力

## 参考文献

1. https://pkg.go.dev/golang.org/x/sync/errgroup