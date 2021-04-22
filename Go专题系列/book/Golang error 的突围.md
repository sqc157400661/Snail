# Golang error 的突围

写过 C 的同学知道，C 语言中常常返回整数错误码（errno）来表示函数处理出错，通常用 `-1` 来表示错误，用 `0` 表示正确。

而在 Go 中，我们使用 `error` 类型来表示错误，不过它不再是一个整数类型，是一个接口类型：

```go
type error interface {
    Error() string
}
```

它表示那些能用一个字符串就能说清的错误。

我们最常用的就是 `errors.New()` 函数，非常简单：

```go
// src/errors/errors.go

func New(text string) error {
	return &errorString{text}
}

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
```

使用 New 函数创建出来的 error 类型实际上是 errors 包里未导出的 `errorString` 类型，它包含唯一的一个字段 `s`，并且实现了唯一的方法：`Error() string`。

通常这就够了，它能反映当时“出错了”，但是有些时候我们需要更加具体的信息，例如：

```go
func Sqrt(f float64) (float64, error) {
    if f < 0 {
        return 0, errors.New("math: square root of negative number")
    }
    // implementation
}
```

当调用者发现出错的时候，只知道传入了一个负数进来，并不清楚到底传的是什么值。在 Go 里：

> It is the error implementation’s responsibility to summarize the context.



它要求返回这个错误的函数要给出具体的“上下文”信息，也就是说，在 `Sqrt` 函数里，要给出这个负数到底是什么。

所以，如果发现 `f` 小于 0，应该这样返回错误：

```go
if f < 0 {
    return 0, fmt.Errorf("math: square root of negative number %g", f)
}
```

这就用到了 `fmt.Errorf` 函数，它先将字符串格式化，再调用 `errors.New` 函数来创建错误。

当我们想知道错误类型，并且打印错误的时候，直接打印 error：

```go
fmt.Println(err)
// 或者
fmt.Println(err.Error)

//注意： fmt 包会自动调用 err.Error() 函数来打印字符串。
```

通常，我们将 error 放到函数返回值的最后一个，没什么好说的，大家都这样做，约定俗成。

参考资料【Tony Bai】这篇文章提到，构造 error 的时候，要求传入的字符串首字母小写，结尾不带标点符号，这是因为我们经常会这样使用返回的 error：

```
... err := errors.New("error example")
fmt.Printf("The returned error is %s.\n", err)
```

# error 的困局

> In Go, error handling is important. The language’s design and conventions encourage you to explicitly check for errors where they occur (as distinct from the convention in other languages of throwing exceptions and sometimes catching them).

在 Go 语言中，错误处理是非常重要的。它从语言层面要求我们需要明确地处理遇到的错误。而不是像其他语言，类如 Java，使用 `try-catch- finally` 这种“把戏”。

这就造成代码里 “error” 满天飞，显得非常冗长拖沓。

而为了代码健壮性考虑，对于函数返回的每一个错误，我们都不能忽略它。因为出错的同时，很可能会返回一个 `nil` 类型的对象。如果不对错误进行判断，那下一行对 `nil` 对象的操作百分之百会引发一个 `panic`。

这样，Go 语言中诟病最多的就是它的错误处理方式似乎回到了上古 C 语言时代。

```go
rr := doStuff1()
if err != nil {
    //handle error...
}

err = doStuff2()
if err != nil {
    //handle error...
}

err = doStuff3()
if err != nil {
    //handle error...
}
```

> Go authors 之一的 Russ Cox 对于这种观点进行过驳斥：当初选择返回值这种错误处理机制而不是 try-catch，主要是考虑前者适用于大型软件，后者更适合小程序。

在参考资料【Go FAQ】里也提到，`try-catch` 会让代码变得非常混乱，程序员会倾向将一些常见的错误，例如，`failing to open a file`，也抛到异常里，这会让错误处理更加冗长繁琐且易出错。

而 Go 语言的多返回值使得返回错误异常简单。对于真正的异常，Go 提供 `panic-recover` 机制，也使得代码看起来非常简洁。

> 当然 Russ Cox 也承认 Go 的错误处理机制对于开发人员的确有一定的心智负担。

参考资料【Go 语言的错误处理机制是一个优秀的设计吗？】是知乎上的一个回答，阐述了 Go 对待错误和异常的不同处理方式，前者使用 error，后者使用 panic，这样的处理比较 Java 那种错误异常一锅端的做法更有优势。

【如何优雅的在Golang中进行错误处理】对于在业务上如何处理 error，给出了一些很好的示例。



# 尝试破局

这部分的内容主要来自 Dave cheney GoCon 2016 的演讲，参考资料可以直达原文。

经常听到 Go 有很多“箴言”，说得很顺口，但理解起来并不是太容易，因为它们大部分都是有故事的。例如，我们常说：

> Don’t communicating by sharing memory, share memory by communicating.

文中还列举了很多，都很有意思：

![go proverbs](D:\www\Snail\Go专题系列\book\images\64576578-d355e180-d3ab-11e9-8e99-34222a3994f1.png)

下面我们讲三条关于 error 的“箴言”。

## Errors are just values

`Errors are just values` 的实际意思是只要实现了 `Error` 接口的类型都可以认为是 `Error`，重要的是要理解这些“箴言”背后的道理。

作者把处理 error 的方式分为三种：

> 1. Sentinel errors
> 2. Error Types
> 3. Opaque errors

我们来挨个说。首先 `Sentinel errors`，Sentinel 来自计算机中常用的词汇，中文意思是“哨兵”。以前在学习快排的时候，会有一个“哨兵”，其他元素都要和“哨兵”进行比较，它划出了一条界限。

这里 `Sentinel errors` 实际想说的是这里有一个错误，暗示处理流程不能再进行下去了，必须要在这里停下，这也是一条界限。**而这些错误，往往是提前约定好的。**

例如，`io` 包里的 `io.EOF`，表示“文件结束”错误。但是这种方式处理起来，不太灵活：

```go
func main() {
	r := bytes.NewReader([]byte("0123456789"))
	
	_, err := r.Read(make([]byte, 10))
	if err == io.EOF {
		log.Fatal("read failed:", err)
	}
}
```

必须要判断 `err` 是否和约定好的错误 `io.EOF` 相等。

再来一个例子，当我想返回 err 并且加上一些上下文信息时，就麻烦了：

```
func main() {
	err := readfile(“.bashrc”)
	if strings.Contains(error.Error(), "not found") {
		// handle error
	}
}

func readfile(path string) error {
	err := openfile(path)
	if err != nil {
		return fmt.Errorf(“cannot open file: %v", err)
	}
	// ……
}
```

在 `readfile` 函数里判断 err 不为空，则用 fmt.Errorf 在 err 前加上具体的 `file` 信息，返回给调用者。返回的 err 其实还是一个字符串。

造成的后果时，调用者不得不用字符串匹配的方式判断底层函数 `readfile` 是不是出现了某种错误。当你必须要这样才能判断某种错误时，代码的“坏味道”就出现了。

顺带说一句，`err.Error()` 方法是给程序员而非代码设计的，也就是说，当我们调用 `Error` 方法时，结果要写到文件或是打印出来，是给程序员看的。在代码里，我们不能根据 `err.Error()` 来做一些判断，就像上面的 `main` 函数里做的那样，不好。

`Sentinel errors` 最大的问题在于它在定义 error 和使用 error 的包之间建立了依赖关系。比如要想判断 `err == io.EOF` 就得引入 io 包，当然这是标准库的包，还 Ok。如果很多用户自定义的包都定义了错误，那我就要引入很多包，来判断各种错误。麻烦来了，这容易引起循环引用的问题。

因此，我们应该尽量避免 `Sentinel errors`，仅管标准库中有一些包这样用，但建议还是别模仿。

第二种就是 `Error Types`，它指的是实现了 `error` 接口的那些类型。它的一个重要的好处是，类型中除了 error 外，还可以附带其他字段，从而提供额外的信息，例如出错的行数等。

标准库有一个非常好的例子：

```go
// PathError records an error and the operation and file path that caused it.
type PathError struct {
	Op   string
	Path string
	Err  error
}
```

`PathError` 额外记录了出错时的文件路径和操作类型。

通常，使用这样的 error 类型，外层调用者需要使用类型断言来判断错误：

```go
// underlyingError returns the underlying error for known os error types.
func underlyingError(err error) error {
	switch err := err.(type) {
	case *PathError:
		return err.Err
	case *LinkError:
		return err.Err
	case *SyscallError:
		return err.Err
	}
	return err
}
```

但是这又不可避免地在定义错误和使用错误的包之间形成依赖关系，又回到了前面的问题。

即使 `Error types` 比 `Sentinel errors` 好一些，因为它能承载更多的上下文信息，但是它仍然存在引入包依赖的问题。因此，也是不推荐的。至少，不要把 `Error types` 作为一个导出类型。

最后一种，`Opaque errors`。翻译一下，就是“黑盒 errors”，因为你能知道错误发生了，但是不能看到它内部到底是什么。

譬如下面这段伪代码：

```go
func fn() error {
	x, err := bar.Foo()
	if err != nil {
		return err
	}
	
	// use x
	return nil
}
```

作为调用者，调用完 `Foo` 函数后，只用知道 `Foo` 是正常工作还是出了问题。也就是说你只需要判断 err 是否为空，如果不为空，就直接返回错误。否则，继续后面的正常流程，不需要知道 err 到底是什么。

这就是处理 `Opaque errors` 这种类型错误的策略。

当然，在某些情况下，这样做并不够用。例如，在一个网络请求中，需要调用者判断返回的错误类型，以此来决定是否重试。这种情况下，作者给出了一种方法：

> In this case rather than asserting the error is a specific type or value, we can assert that the error implements a particular behaviour.

就是说，**不去判断错误的类型到底是什么，而是去判断错误是否具有某种行为**，**或者说实现了某个接口**。

来个例子：

```go
type temporary interface {
	Temporary() bool
}

func IsTemporary(err error) bool {
	te, ok := err.(temporary)
	return ok && te.Temporary()
}
```

拿到网络请求返回的 error 后，调用 `IsTemporary` 函数，如果返回 true，那就重试。

这么做的好处是在进行网络请求的包里，不需要 `import` 引用定义错误的包。

## handle not just check errors

这一节要说第二句箴言：“Don’t just check errors, handle them gracefully”。

```go
func AuthenticateRequest(r *Request) error {
     err := authenticate(r.User)
     if err != nil {
        return err
     }
     return nil
}
```

上面这个例子中的代码是有问题的，直接优化成一句就可以了：

```go
func AuthenticateRequest(r *Request) error {
     return authenticate(r.User)
}
```

还有其他的问题，在函数调用链的最顶层，我们得到的错误可能是：`No such file or directory`。

这个错误反馈的信息太少了，不知道文件名、路径、行号等等。

尝试改进一下，增加一点上下文：

```go
func AuthenticateRequest(r *Request) error {
     err := authenticate(r.User)
     if err != nil {
        return fmt.Errorf("authenticate failed: %v", err)
     }
     return nil
}
```

这种做法实际上是先错误转换成字符串，再拼接另一个字符串，最后，再通过 `fmt.Errorf` 转换成错误。这样做破坏了相等性检测，即我们无法判断错误是否是一种预先定义好的错误了。

应对方案是使用第三方库：`github.com/pkg/errors`。提供了友好的界面：

```go
// Wrap annotates cause with a message.
func Wrap(cause error, message string) error
// Cause unwraps an annotated error.
func Cause(err error) error
```

通过 `Wrap` 可以将一个错误，加上一个字符串，“包装”成一个新的错误；通过 `Cause` 则可以进行相反的操作，将里层的错误还原。

有了这两个函数，就方便很多：

```go
func ReadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "open failed")
	}
	defer f.Close()
	
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "read failed")
	}
	return buf, nil
}
```

当在外层调用 `ReadFile` 函数时：

```go
func main() {
	_, err := ReadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ReadConfig() ([]byte, error) {
	home := os.Getenv("HOME")
	config, err := ReadFile(filepath.Join(home, ".settings.xml"))
	return config, errors.Wrap(err, "could not read config")
}
```

它是有层次的，非常清晰。而如果我们用 `pkg/errors` 库提供的打印函数：

```go
func main() {
	_, err := ReadConfig()
	if err != nil {
		errors.Print(err)
		os.Exit(1)
	}
}
```

能得到更有层次、更详细的错误：

```go
readfile.go:27: could not read config
readfile.go:14: open failed
open /Users/dfc/.settings.xml: no such file or directory
```

上面讲的是 `Wrap` 函数，接下来看一下 “Cause” 函数，以前面提到的 `temporary` 接口为例：

```go
type temporary interface {
	Temporary() bool
}

// IsTemporary returns true if err is temporary.
func IsTemporary(err error) bool {
	te, ok := errors.Cause(err).(temporary)
	return ok && te.Temporary()
}
```

判断之前先使用 `Cause` 取出错误，做断言，最后，递归地调用 `Temporary` 函数。如果错误没实现 `temporary` 接口，就会断言失败，返回 `false`。

## Only handle errors once

什么叫“处理”错误：

> Handling an error means inspecting the error value, and making a decision.

意思是查看了一下错误，并且做出一个决定。

例如，如果不做任何决定，相当于忽略了错误：

```go
func Write(w io.Writer, buf []byte) {
	w.Write(buf)
}
```

`w.Write(buf)` 会返回两个结果，一个表示写成功的字节数，一个是 error，上面的例子中没有对这两个返回值做任何处理。

下面这个例子却又处理了两次错误：

```go
func Write(w io.Writer, buf []byte) error {
	_, err := w.Write(buf)
	if err != nil {
		// annotated error goes to log file
		log.Println("unable to write:", err)
	
		// unannotated error returned to caller return err
		return err
	}
	return nil
}
```

第一次处理是将错误写进了日志，第二次处理则是将错误返回给上层调用者。而调用者也可能将错误写进日志或是继续返回给上层。

这样一来，日志文件中会有很多重复的错误描述，并且在最上层调用者（如 main 函数）看来，它拿到的错误却还是最底层函数返回的 error，没有任何上下文信息。

使用第三方的 error 包就可以比较完美的解决问题：

```go
func Write(w io.Write, buf []byte) error {
	_, err := w.Write(buf)
	return errors.Wrap(err, "write failed")
}
```

返回的错误，对于人和机器而言，都是友好的。

## 小结

这一部分主要讲了处理 error 的一些原则，引入了第三方的 errors 包，使得错误处理变得更加优雅。

作者最后给出了一些结论：

> 1. errors 就像对外提供的 API 一样，需要认真对待。
> 2. 将 errors 看成黑盒，判断它的行为，而不是类型。
> 3. 尽量不要使用 sentinel errors。
> 4. 使用第三方的错误包来包裹 error（errors.Wrap），使得它更好用。
> 5. 使用 errors.Cause 来获取底层的错误。

# 胎死腹中的 try 提案

之前已经出现用 “check & handle” 关键字和 “try 内置函数”改进错误处理流程的提案，目前 try 内置函数的提案已经被官方提前拒绝，原因是社区里一边倒地反对声音。

关于这两个提案的具体内容见参考资料【check & handle】和【try 提案】。

# go 1.13 的改进

有一些 Go 语言失败的尝试，比如 Go 1.5 引入的 vendor 和 internal 来管理包，最后被滥用而引发了很多问题。因此 Go 1.13 直接抛弃了 `GOPATH` 和 `vendor` 特性，改用 `module` 来管理包。

柴大在《Go 语言十年而立，Go2 蓄势待发》一文中表示：

> 比如最近 Go 语言之父之一 Robert Griesemer 提交的通过 try 内置函数来简化错误处理就被否决了。失败的尝试是一个好的现象，它表示 Go 语言依然在一些新兴领域的尝试 —— Go 语言依然处于活跃期。

2019年 9 月 3 号，Go 发布 1.13 版本，除了 module 特性转正之外，还改进了数字字面量。比较重要的还有 defer 性能提升 30%，将更多的对象从堆上移动到栈上以提升性能，等等。

> 还有一个重大的改进发生在 errors 标准库中。errors 库增加了 Is/As/Unwrap三个函数，这将用于支持错误的再次包装和识别处理，为 Go 2 中新的错误处理改进提前做准备。

`1.13` 支持了 `error` 包裹（wrapping）：

> An error e can wrap another error w by providing an Unwrap method that returns w. Both e and w are available to programs, allowing e to provide additional context to w or to reinterpret it while still allowing programs to make decisions based on w.

为了支持 wrapping，`fmt.Errorf` 增加了 `%w` 的格式，并且在 `error` 包增加了三个函数：`errors.Unwrap`，`errors.Is`，`errors.As`。

## fmt.Errorf

使用 `fmt.Errorf` 加上 `%w` 格式符来生成一个嵌套的 error，它并没有像 `pkg/errors` 那样使用一个 Wrap 函数来嵌套 error，非常简洁。

## Unwrap

```go
func Unwrap(err error) error
```

将嵌套的 error 解析出来，多层嵌套需要调用 `Unwrap` 函数多次，才能获取最里层的 error。

源码如下：

```go
func Unwrap(err error) error {
    // 判断是否实现了 Unwrap 方法
	u, ok := err.(interface {
		Unwrap() error
	})
	// 如果不是，返回 nil
	if !ok {
		return nil
	}
	// 调用 Unwrap 方法返回被嵌套的 error
	return u.Unwrap()
}
```

对 err 进行断言，看它是否实现了 Unwrap 方法，如果是，调用它的 Unwrap 方法。否则，返回 nil。

## Is

```go
func Is(err, target error) bool
```

判断 err 是否和 target 是同一类型，或者 err 嵌套的 error 有没有和 target 是同一类型的，如果是，则返回 true。

源码如下：

```go
func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	isComparable := reflectlite.TypeOf(target).Comparable()
	
	// 无限循环，比较 err 以及嵌套的 error
	for {
		if isComparable && err == target {
			return true
		}
		// 调用 error 的 Is 方法，这里可以自定义实现
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		// 返回被嵌套的下一层的 error
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}
```

通过一个无限循环，使用 `Unwrap` 不断地将 err 里层嵌套的 error 解开，再看被解开的 error 是否实现了 Is 方法，并且调用它的 Is 方法，当两者都返回 true 的时候，整个函数返回 true。

## As

```go
func As(err error, target interface{}) bool
```

从 err 错误链里找到和 target 相等的并且设置 target 所指向的变量。

源码如下：

```go
func As(err error, target interface{}) bool {
    // target 不能为 nil
	if target == nil {
		panic("errors: target cannot be nil")
	}
	
	val := reflectlite.ValueOf(target)
	typ := val.Type()
	
	// target 必须是一个非空指针
	if typ.Kind() != reflectlite.Ptr || val.IsNil() {
		panic("errors: target must be a non-nil pointer")
	}
	
	// 保证 target 是一个接口类型或者实现了 Error 接口
	if e := typ.Elem(); e.Kind() != reflectlite.Interface && !e.Implements(errorType) {
		panic("errors: *target must be interface or implement error")
	}
	targetType := typ.Elem()
	for err != nil {
	    // 使用反射判断是否可被赋值，如果可以就赋值并且返回true
		if reflectlite.TypeOf(err).AssignableTo(targetType) {
			val.Elem().Set(reflectlite.ValueOf(err))
			return true
		}
		
		// 调用 error 自定义的 As 方法，实现自己的类型断言代码
		if x, ok := err.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}
		// 不断地 Unwrap，一层层的获取嵌套的 error
		err = Unwrap(err)
	}
	return false
}
```

返回 true 的条件是错误链里的 err 能被赋值到 target 所指向的变量；或者 err 实现的 `As(interface{}) bool` 方法返回 true。

前者，会将 err 赋给 target 所指向的变量；后者，由 As 函数提供这个功能。

如果 target 不是一个指向“实现了 error 接口的类型或者其它接口类型”的非空的指针的时候，函数会 panic。

这一部分的内容，飞雪无情大佬的文章【飞雪无情 分析 1.13 错误】写得比较好，推荐阅读。

# 总结

Go 语言使用 error 和 panic 处理错误和异常是一个非常好的做法，比较清晰。至于是使用 error 还是 panic，看具体的业务场景。

当然，Go 中的 error 过于简单，以至于无法记录太多的上下文信息，对于错误包裹也没有比较好的办法。当然，这些可以通过第三方库来解决。官方也在新发布的 go 1.13 中对这一块作出了改进，相信在 Go 2 里会有更进一步的优化。

本文还列举了一些处理 error 的示例，例如不要两次处理一个错误，判断错误的行为而不是类型等等。

参考资料里列举了很多错误处理相关的示例，这篇文章作为一个引子。

# 参考资料

【Go 2 错误提案】https://go.googlesource.com/proposal/+/master/design/29934-error-values.md

【check & handle】https://go.googlesource.com/proposal/+/master/design/go2draft-error-handling-overview.md

【错误讨论的 issue】https://github.com/golang/go/issues/29934

【error value 的 FAQ】https://github.com/golang/go/wiki/ErrorValueFAQ

【error 包】https://golang.org/pkg/errors/

【飞雪无情的博客 错误处理】https://www.flysnow.org/2019/01/01/golang-error-handle-suggestion.html

【飞雪无情 分析 1.13 错误】https://www.flysnow.org/2019/09/06/go1.13-error-wrapping.html

【Tony Bai Go语言错误处理】https://tonybai.com/2015/10/30/error-handling-in-go/

【Go 官方 error 使用教程】https://blog.golang.org/error-handling-and-go

【Go FAQ】https://golang.org/doc/faq#exceptions

【ethancai 错误处理】https://ethancai.github.io/2017/12/29/Error-Handling-in-Go/

【Dave cheney GoCon 2016 演讲】https://dave.cheney.net/paste/gocon-spring-2016.pdf

【Morsing’s Blog Effective error handling in Go】http://morsmachine.dk/error-handling

【如何优雅的在Golang中进行错误处理】https://www.ituring.com.cn/article/508191

【Go 2 错误处理提案：try 还是 check？】https://toutiao.io/posts/uh9qo7/preview

【try 提案】https://github.com/golang/go/issues/32437

【否决 try 提案】https://github.com/golang/go/issues/32437#issuecomment-512035919

【Go 语言的错误处理机制是一个优秀的设计吗？】https://www.zhihu.com/question/27158146/answer/44676012