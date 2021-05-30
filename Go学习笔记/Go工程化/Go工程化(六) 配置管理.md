# Go工程化(六) 配置管理

## 序

**3 月进度: 03/15** 3 月开始会尝试爆更模式，争取做到两天更新一篇文章，如果感兴趣可以拉到文章最下方获取关注方式。

应用的配置大概可以分为以下几种

- 环境配置
  - 环境配置，应该是应用部署时就已经确定好的信息，这些信息不应该写在我们的配置文件或者是放到配置中心，而是应该由我们的部署平台，例如 K8s 直接在容器启动时候就注入好
  - region: 区域信息
  - env: 环境信息，例如 prod, test
  - zone: 可用区
  - host: 机器名
  - appid: 应用 id
  - color: 流量染色信息，用来做流量分发的
- 静态配置
  - 资源需要初始化的配置信息，比如 http/gRPC server、redis、mysql 等
  - 这类资源在线变更配置的风险非常大，尽量不要在线动态变更，很可能会导致业务出现不可预期的事故
  - 变更静态配置和发布 bianry app 没有区别，应该走一次迭代发布的流程。
- 动态配置
  - 应用程序可能需要一些在线的开关，来控制业务的一些简单策略，会频繁的调整和使用，我们把这类是基础类型(int, bool)等配置，用于可以动态变更业务流的收归一起，
  - 不过业务配置最好做到管理后台，因为配置中心运营同学一般没有权限，并且很多配置中心的校验做的不够好，不熟悉的人进行变更很容易出问题
- 全局配置
  - 我们依赖的各类组件、中间件都有大量的默认配置或者指定配置，在各个项目里大量拷贝复制，容易出现意外，所以我们使用全局配置模板来定制化常用的组件，然后再特化的应用里进行局部替换。

## 函数参数配置

下面这个是 redis 初始化的例子，一般在我们刚刚开始写代码的时候，我们都会向下面这么写，把需要的参数放到函数的入参就行了。

```go
func Dial(network, address string) (Conn, error)
```

这个有什么问题呢？如果这个函数只是你自己用也没有什么毛病，但是如果是一个公共的库或者是中间件就会发现，用户的诉求是多种多样，灵活多变的。就会听见

- 我要自定义超时时间
- 我要自定义 database

等等一系列的需求和各种各样的声音。
这时候为了满足大家的需求，最简单，最直接的做法就是，为不同的需求添加不同的初始化函数

```go
func DialTimeout(network, address string,
                 connectTimeout, readTimeout, writeTimeout time.Duration) (Conn, error)

func DialDatabase(network, address string, database int) (Conn, error)
```

但这样毕竟不是一个办法，因为用户的需求是满足不完的，作为公共库，不可能为每个用户的需求都单独来搞个函数签名，那这样函数签名也太多了。而且还有一个问题是参数列表会很长，例如上面的 `DialTimeout` , 可读性也不好。
当然这也和 Go 的函数不能重载有关系，如果可以重载的话，每种需求来一个可能也还行，但是其实也不够优雅。

这时候我们比较容易想到的办法是什么呢？既然参数比较长，配置变化又想要灵活，那么我们就直接传入一个对象就好了，让每个用户自己构造去。

```go
type Config struct {
  *pool.Config
  Addr string
  Auth string
  DialTimeout time.Duration
  ReadTimeout time.Duration
  WriteTimeout time.Duration
}

// NewConn new a redis conn.
func NewConn(c *Config) (cn Conn, err error)
```

这种方式有什么问题呢？

- 可以看到 `NewConn` 传递的是一个指针，那么这个只能就能够被外面修改，只要外面修改那就麻烦了，因为不知道会发生什么，这是一个未定义的行为。
- 还有就是我们没有办法指定必填参数，这样传递相当于每一项都是可选的

既然指针可能会导致未定义的行为，那我们就换个方式, 不传指针传结构体不就行了

```go
func NewConn(c Config) (cn Conn, err error)
```

但是这又带来了一些新的问题

- 首先，必填参数的问题还是没有解决的
- 其次，这么传参我们是没有办法区分默认值的，通过指针我们可以通过判断是否等于 nil 来区分，因为大部分的场景下其实用默认值就可以了，这样做反而降低了使用体验

所以，有一段时间毛老师他们都是使用上面传指针的这种方式，当然这种方式我们也用过，虽然可以用，但是就是有点不爽

> “I believe that we, as Go programmers, should work hard to ensure that nil is never a parameter that needs to be passed to any public function.” – Dave Cheney

dava 大神也提到过，我们应该将 nil 作为一个函数的参数值进行传递，那我们该如何修改呢？
如果去看一些知名的开源库或者是标准库的一些初始化代码，我们可以看到这种姿势

```go
type DialOption struct {
  f func(*dialOptions)
}

func Dial(network, address string, options ...DialOption) (Conn, error) {
  do := dialOptions{
    dial: net.Dial,
  }
  for _, option := range options {
    option.f(&do)
  } // ...
}

// DialReadTimeout specifies the timeout for reading a single command reply.
func DialReadTimeout(d time.Duration) DialOption {
	return DialOption{func(do *dialOptions) {
		do.readTimeout = d
	}}
}
```

这种操作的核心在于，我们可以定义一个未导出的 `option struct` 用于存放配置，然后导出一个函数指针，然后我们在初始化的时候，使用可变参数进行传递，然后再初始化函数内部通过 for 循环调用修改相关的配置。

- 这样我们就可以把必填参数放在前面几位，保证参数必填，一眼就能看出来，减少沟通成本
- 然后默认参数，我们可以在函数内部先初始化一个 defaultOption 然后用后面配置的函数进行修改即可

我们可以在包里面直接定义一些函数例如上面的 `DialReadTimeout` 来返回一个函数，然后进行配置修改

但是这样就可以了么？这种使用方式

- 首先，函数指针没有必要搞那么麻烦，其实直接顶一个函数类型就可以了 `type DialOption func(*dialOptions)`
- 其次，这种做法还是只能在包内部进行定义，用户是没有办法自定义一些配置的，但是其实也够用了

如果想要用户可以自定义一些配置，我们可以看看 grpc 的配置定义，主要的思路就是把 option 从函数修改接口，然后定义了一个 `EmptyCallOption` 实现这个接口，因为这个接口包含的函数是未导出的，所以我们只要在需要做配置的 struct 当中包含这个 `EmptyCallOption` 就可以了

```go
type GreeterClient interface {
  SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption)     (*HelloReply, error)
}

type CallOption interface {
  before(*callInfo) error
  after(*callInfo)
}
// EmptyCallOption does not alter the Call configuration.
type EmptyCallOption struct{}

// TimeoutCallOption timeout option.
type TimeoutCallOption struct {
  grpc.EmptyCallOption
  Timeout time.Duration
}
```

## 配置文件

到这里函数的配置就解决了，但是我们怎么和配置文件进行结合呢？现在的这种做法隐藏了结构，没有办法直接使用 `json.Unmarshal` 这种方法直接反序列化回来。

比较常见的办法就是我们设定两个函数如果需要配置文件反序列化的就用不带 Option 的，反之用带 Option 的

```go
func Dial(network, address string, options ...DialOption) (Conn, error)

// NewConn new a redis conn.
func NewConn(c *Config) (cn Conn, err error)
```

这么做比较大的问题就是，把 config 给暴露了出来，并且有两种初始化方式，使用配置文件就没有办法得到使用 Option 的好处了

课上提供了一种解决思路就是把这两步进行分离，首先我们使用 protobuf 文件定义好配置的结构，这样可以加上一些验证条件

```go
syntax = "proto3";
import "google/protobuf/duration.proto";
package config.redis.v1;
// redis config.
message redis {
  string network = 1;
  string address = 2;
  int32 database = 3;
  string password = 4;
  google.protobuf.Duration read_timeout = 5;
}
```

定义好之后使用 yaml 来修改配置，然后使用 `Options` 方法，将 protobuf 生成的 `Config` 替换为 redis.Options

```go
func ApplyYAML(s *redis.Config, yml string) error {
  js, err := yaml.YAMLToJSON([]byte(yml))
  if err != nil {
    return err
  }
  return ApplyJSON(s, string(js))
}
// Options apply config to options.
func Options(c *redis.Config) []redis.Options {
  return []redis.Options{
    redis.DialDatabase(c.Database),
    redis.DialPassword(c.Password),
    redis.DialReadTimeout(c.ReadTimeout),
  }
}
```

这种方式除了定义起来比较麻烦，使用上还是很简单的，使用只需要像下面这样就可以了

```go
func main() {
  // load config file from yaml.
  c := new(redis.Config)
  _ = ApplyYAML(c, loadConfig())
  r, _ := redis.Dial(c.Network, c.Address, Options(c)...)
}
```

由于我们现在使用的没有那么复杂，统一接入了配置中心，所以我现在的做法是定义一个 `WithConfigCenter` 的方法就行了，调用的时候其实还要简单一点

```go
func WithConfigCenter(config ConfigCenter, key string) Option
```

## 总结

修改配置其实是一件比较危险的事情，很多时候我们缺乏足够的敬畏，因为现在在线的配置中心越来方便，所以修改的成本越来越低，大家就越来越随意，所以我们需要对配置的修改慎重一些。配置的目标：

- 避免复杂
- 多样的配置
- 简单化努力
- 以基础设施 -> 面向用户进行转变
- 配置的必选项和可选项
- 配置的防御编程
- 权限和变更跟踪
- 配置的版本和应用对齐，这个很多都没做到，经常应用回滚了配置没回滚，就出事故了
- 安全的配置变更：逐步部署、回滚更改、自动回滚

## 参考文献

1. [Go 进阶训练营-极客时间](https://u.geekbang.org/subject/go?utm_source=lailin.xyz&utm_medium=lailin.xyz)
2. [command center: Self-referential functions and the design of options](https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html)
3. [Functional options for friendly APIs – The acme of foolishness](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)