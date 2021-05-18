# Go常见语法面试题

### 常⻅语法题目 一

### 1、下面代码能运行吗？为什么。

27 for i := 0; i < repeat; i++ {
28 x, y, z = move(repeatCmd, x, y, z)

(^29) }
(^30) repeat = 0
(^31) repeatCmd = ""
(^32) case repeat > 0 && s != '(' && s != ')':
(^33) repeatCmd = repeatCmd + string(s)
34 case s == 'L':
35 z = (z + 1) % 4
36 case s == 'R':
(^37) z = (z - 1 + 4) % 4
(^38) case s == 'F':
(^39) switch {
(^40) case z == Left || z == Right:
(^41) x = x - z + 1
(^42) case z == Top || z == Bottom:
43 y = y - z + 2
44 }
45 case s == 'B':
(^46) switch {
(^47) case z == Left || z == Right:
(^48) x = x + z - 1
(^49) case z == Top || z == Bottom:
(^50) y = y + z - 2
51 }
52 }
53 }
(^54) return
(^55) }
56
1 type Param map[string]interface{}
(^23) type Show struct {
4 Param
5 }
67 func main1() {
(^8) s := new(Show)
(^9) s.Param["RMB"] = 10000


#### 解析

#### 共发现两个问题：

1. main  函数不能加数字。
2. new  关键字无法初始化 Show  结构体中的  Param  属性，所以直接
    对  s.Param  操作会出错。

### 2、请说出下面代码存在什么问题。

#### 解析：

```
golang中有规定，switch type 的case T1，类型列表只有一个，那么 v := m.(type)
中的 v 的类型就是T1类型。
如果是case T1, T2，类型列表中有多个，那 v 的类型还是多对应接口的类型，也就
是m 的类型。
所以这里 msg的类型还是 interface{} ，所以他没有 Name 这个字段，编译阶段就会
报错。具体解释⻅： https://golang.org/ref/spec#Type_switches
```
### 3、写出打印的结果。

10 }

```
1 type student struct {
```
(^2) Name string
(^3) }
(^45) func zhoujielun(v interface{}) {
6 switch msg := v.(type) {
7 case *student, student:
8 msg.Name
(^9) }
(^10) }
1 type People struct {
2 name string `json:"name"`
(^3) }
(^45) func main() {
(^6) js := `{
(^7) "name":"11"
(^8) }`
(^9) var p People
10 err := json.Unmarshal([]byte(js), &p)
11 if err != nil {
12 fmt.Println("err: ", err)
(^13) return
(^14) }
(^15) fmt.Println("people: ", p)
(^16) }


#### 解析：

```
按照 golang 的语法，小写开头的方法、属性或 struct  是私有的，同样，在json  解
码或转码的时候也无法上线私有属性的转换。
题目中是无法正常得到 People 的name 值的。而且，私有属性name 也不应该加
json的标签。
```
### 4、下面的代码是有问题的，请说明原因。

#### 解析：

```
在golang中 String() string 方法实际上是实现了 String 的接口的，该接口定义在
fmt/print.go  中：
```
```
在使用  fmt  包中的打印方法时，如果类型实现了这个接口，会直接调用。而题目中打
印  p 的时候会直接调用 p 实现的  String() 方法，然后就产生了循环调用。
```
### 5、请找出下面代码的问题所在。

```
1 type People struct {
2 Name string
3 }
```
(^45) func (p *People) String() string {
(^6) return fmt.Sprintf("print: %v", p)
(^7) }
(^89) func main() {
(^10) p := &People{}
11 p.String()
12 }
1 type Stringer interface {
(^2) String() string
(^3) }
1 func main() {
(^2) ch := make(chan int, 1000)
(^3) go func() {
(^4) for i := 0; i < 10; i++ {
5 ch <- i
6 }
(^7) }()
(^8) go func() {
(^9) for {
(^10) a, ok := <-ch
(^11) if !ok {
(^12) fmt.Println("close")
13 return
14 }


#### 解析：

```
在 golang 中 goroutine  的调度时间是不确定的，在题目中，第一个
写  channel 的  goroutine  可能还未调用，或已调用但没有写完时直接  close  管道，
可能导致写失败，既然出现 panic  错误。
```
### 6、请说明下面代码书写是否正确。

#### 解析：

```
atomic.CompareAndSwapInt32  函数不需要循环调用。
```
### 7、下面的程序运行后为什么会爆异常。

15 fmt.Println("a: ", a)
16 }

(^17) }()
(^18) close(ch)
(^19) fmt.Println("ok")
(^20) time.Sleep(time.Second * 100)
(^21) }
1 var value int
23 func SetValue(delta int32) {
(^4) for {
(^5) v := value
(^6) if atomic.CompareAndSwapInt32(&value, v, (v+delta)) {
(^7) break
(^8) }
(^9) }
10 }
1 type Project struct{}
23 func (p *Project) deferError() {
4 if err := recover(); err != nil {
(^5) fmt.Println("recover: ", err)
(^6) }
(^7) }
(^89) func (p *Project) exec(msgchan chan interface{}) {
(^10) for msg := range msgchan {
11 m := msg.(int)
12 fmt.Println("msg: ", m)
13 }
(^14) }
(^1516) func (p *Project) run(msgchan chan interface{}) {
(^17) for {
(^18) defer p.deferError()
(^19) go p.exec(msgchan)


#### 解析：

#### 有一下几个问题：

1. time.Sleep 的参数数值太大，超过了  1<<63 - 1  的限制。
2. defer p.deferError()  需要在协程开始出调用，否则无法捕获 panic 。

### 8、请说出下面代码哪里写错了

#### 解析：

#### 协程可能还未启动，管道就关闭了。

### 9、请说出下面代码，执行时为什么会报错

20 time.Sleep(time.Second * 2)
21 }

(^22) }
(^2324) func (p *Project) Main() {
(^25) a := make(chan interface{}, 100)
(^26) go p.run(a)
(^27) go func() {
28 for {
29 a <- "1"
30 time.Sleep(time.Second)
(^31) }
(^32) }()
(^33) time.Sleep(time.Second * 100000000000000)
(^34) }
(^3536) func main() {
(^37) p := new(Project)
38 p.Main()
39 }
1 func main() {
(^2) abc := make(chan int, 1000)
(^3) for i := 0; i < 10; i++ {
(^4) abc <- i
(^5) }
6 go func() {
7 for a := range abc {
8 fmt.Println("a: ", a)
(^9) }
(^10) }()
(^11) close(abc)
(^12) fmt.Println("close")
(^13) time.Sleep(time.Second * 100)
(^14) }


#### 解析：

```
map的value本身是不可寻址的，因为map中的值会在内存中移动，并且旧的指针地址在
map改变时会变得无效。故如果需要修改map值，可以将 map 中的非指针类型
value，修改为指针类型，比如使用 map[string]*Student.
```
### 10、请说出下面的代码存在什么问题？

#### 解析：

```
依据4个goroutine的启动后执行效率，很可能打印111func4，但其他的111func*也
可能先执行，exec只会返回一条信息。
```
### 11、下面这段代码为什么会卡死？

```
1 type Student struct {
2 name string
3 }
45 func main() {
```
(^6) m := map[string]Student{"people": {"zhoujielun"}}
(^7) m["people"].name = "wuyanzu"
(^8) }
1 type query func(string) string
(^23) func exec(name string, vs ...query) string {
4 ch := make(chan string)
5 fn := func(i int) {
6 ch <- vs[i](name)
(^7) }
(^8) for i, _ := range vs {
(^9) go fn(i)
(^10) }
(^11) return <-ch
12 }
1314 func main() {
15 ret := exec("111", func(n string) string {
(^16) return n + "func1"
(^17) }, func(n string) string {
(^18) return n + "func2"
(^19) }, func(n string) string {
(^20) return n + "func3"
(^21) }, func(n string) string {
22 return n + "func4"
23 })
24 fmt.Println(ret)
(^25) }
1 package main


#### 解析：

```
Golang 中，byte 其实被 alias 到 uint8 上了。所以上面的 for 循环会始终成立，因为
i++ 到 i=255 的时候会溢出，i <= 255 一定成立。
也即是， for 循环永远无法退出，所以上面的代码其实可以等价于这样：
```
```
正在被执行的 goroutine 发生以下情况时让出当前 goroutine 的执行权，并调度后面的
goroutine 执行：
 IO 操作
 Channel 阻塞
 system call
 运行较⻓时间
如果一个 goroutine 执行时间太⻓，scheduler 会在其 G 对象上打上一个标志（
preempt），当这个 goroutine 内部发生函数调用的时候，会先主动检查这个标志，如
果为 true 则会让出执行权。
main 函数里启动的 goroutine 其实是一个没有 IO 阻塞、没有 Channel 阻塞、没有
system call、没有函数调用的死循环。
也就是，它无法主动让出自己的执行权，即使已经执行很⻓时间，scheduler 已经标志
了 preempt。
而 golang 的 GC 动作是需要所有正在运行 goroutine  都停止后进行的。因此，程序
会卡在  runtime.GC() 等待所有协程退出。
```
### 常⻅语法题目 二

### 1、写出下面代码输出内容。

```
23 import (
4 "fmt"
```
(^5) "runtime"
(^6) )
(^78) func main() {
(^9) var i byte
(^10) go func() {
11 for i = 0; i <= 255; i++ {
12 }
13 }()
(^14) fmt.Println("Dropping mic")
(^15) // Yield execution to force executing other goroutines
(^16) runtime.Gosched()
(^17) runtime.GC()
(^18) fmt.Println("Done")
(^19) }
1 go func() {
(^2) for {}
(^3) }


#### 解析：

```
defer 关键字的实现跟go关键字很类似，不同的是它调用的是 runtime.deferproc而不
是runtime.newproc 。
在defer 出现的地方，插入了指令 call runtime.deferproc，然后在函数返回之前的地
方，插入指令call runtime.deferreturn 。
goroutine的控制结构中，有一张表记录defer ，调用runtime.deferproc 时会将需要
defer的表达式记录在表中，而在调用 runtime.deferreturn 的时候，则会依次从defer表
中出栈并执行。
因此，题目最后输出顺序应该是 defer 定义顺序的倒序。 panic  错误并不能终
止  defer 的执行。
```
### 2、 以下代码有什么问题，说明原因

#### 解析：

```
golang 的  for ... range 语法中， stu 变量会被复用，每次循环会将集合中的值复制
给这个变量，因此，会导致最后 m中的 map 中储存的都是stus 最后一个 student
```
```
1 package main
23 import (
4 "fmt"
5 )
```
(^67) func main() {
(^8) defer_call()
(^9) }
(^1011) func defer_call() {
(^12) defer func() { fmt.Println("打印前") }()
(^13) defer func() { fmt.Println("打印中") }()
14 defer func() { fmt.Println("打印后") }()
1516 panic("触发异常")
17 }
1 type student struct {
(^2) Name string
(^3) Age int
(^4) }
(^56) func pase_student() {
(^7) m := make(map[string]*student)
8 stus := []student{
9 {Name: "zhou", Age: 24},
10 {Name: "li", Age: 23},
(^11) {Name: "wang", Age: 22},
(^12) }
(^13) for _, stu := range stus {
(^14) m[stu.Name] = &stu
(^15) }
(^16) }


#### 的值。

### 3、下面的代码会输出什么，并说明原因

#### 解析：

```
这个输出结果决定来自于调度器优先调度哪个G。从runtime的源码可以看到，当创建一
个G时，会优先放入到下一个调度的runnext 字段上作为下一次优先调度的G。因此，
最先输出的是最后创建的G，也就是9.
```
```
1 func main() {
```
(^2) runtime.GOMAXPROCS(1)
(^3) wg := sync.WaitGroup{}
(^4) wg.Add(20)
5 for i := 0; i < 10; i++ {
6 go func() {
7 fmt.Println("i: ", i)
(^8) wg.Done()
(^9) }()
(^10) }
(^11) for i := 0; i < 10; i++ {
(^12) go func(i int) {
(^13) fmt.Println("i: ", i)
14 wg.Done()
15 }(i)
16 }
(^17) wg.Wait()
(^18) }
1 func newproc(siz int32, fn *funcval) {
2 argp := add(unsafe.Pointer(&fn), sys.PtrSize)
3 gp := getg()
4 pc := getcallerpc()
(^5) systemstack(func() {
(^6) newg := newproc1(fn, argp, siz, gp, pc)
(^78) _p_ := getg().m.p.ptr()
(^9) //新创建的G会调用这个方法来决定如何调度
(^10) runqput(_p_, newg, true)
1112 if mainStarted {
13 wakep()
14 }
(^15) })
(^16) }
(^17) ...
(^1819) if next {
(^20) retryNext:
(^21) oldnext := _p_.runnext
22 //当next是true时总会将新进来的G放入下一次调度字段中
23 if !_p_.runnext.cas(oldnext, guintptr(unsafe.Pointer(gp))) {


### 4、下面代码会输出什么？

#### 解析：

```
输出结果为 showA、 showB。golang 语言中没有继承概念，只有组合，也没有虚方
法，更没有重载。因此， *Teacher  的  ShowB  不会覆写被组合的 People  的方法。
```
### 5、下面代码会触发异常吗？请详细说明

24 goto retryNext
25 }

(^26) if oldnext == 0 {
(^27) return
(^28) }
(^29) // Kick the old runnext out to the regular run queue.
(^30) gp = oldnext.ptr()
31 }
1 type People struct{}
(^23) func (p *People) ShowA() {
(^4) fmt.Println("showA")
5 p.ShowB()
6 }
7 func (p *People) ShowB() {
(^8) fmt.Println("showB")
(^9) }
(^1011) type Teacher struct {
(^12) People
(^13) }
1415 func (t *Teacher) ShowB() {
16 fmt.Println("teacher showB")
17 }
(^1819) func main() {
(^20) t := Teacher{}
(^21) t.ShowA()
(^22) }
1 func main() {
(^2) runtime.GOMAXPROCS(1)
3 int_chan := make(chan int, 1)
4 string_chan := make(chan string, 1)
5 int_chan <- 1
(^6) string_chan <- "hello"
(^7) select {
(^8) case value := <-int_chan:
(^9) fmt.Println(value)
(^10) case value := <-string_chan:


#### 解析：

```
结果是随机执行。golang 在多个 case  可读的时候会公平的选中一个执行。
```
### 6、下面代码输出什么？

#### 解析：

#### 输出结果为：

```
defer 在定义的时候会计算好调用函数的参数，所以会优先输出 10 、 20  两个参
数。然后根据定义的顺序倒序执行。
```
### 7、请写出以下输入内容

#### 解析：

#### 输出为  0 0 0 0 0 1 2 3 。

```
make 在初始化切片时指定了⻓度，所以追加数据时会从 len(s)  位置开始填充数据。
```
11 panic(value)
12 }

(^13) }
1 func calc(index string, a, b int) int {
(^2) ret := a + b
(^3) fmt.Println(index, a, b, ret)
(^4) return ret
(^5) }
67 func main() {
8 a := 1
9 b := 2
(^10) defer calc("1", a, calc("10", a, b))
(^11) a = 0
(^12) defer calc("2", a, calc("20", a, b))
(^13) b = 1
(^14) }
1 10 1 2 3
2 20 0 2 2
3 2 0 2 2
(^4) 1 1 3 4
1 func main() {
(^2) s := make([]int, 5)
(^3) s = append(s, 1, 2, 3)
(^4) fmt.Println(s)
(^5) }


### 8、下面的代码有什么问题?

#### 解析：

```
在执行 Get方法时可能被panic。
虽然有使用sync.Mutex做写锁，但是map是并发读写不安全的。map属于引用类型，并
发读写时多个协程⻅是通过指针访问同一个地址，即访问共享变量，此时同时读写资源
存在竞争关系。会报错误信息:“fatal error: concurrent map read and map write”。
因此，在 Get  中也需要加锁，因为这里只是读，建议使用读写锁  sync.RWMutex 。
```
### 9、下面的迭代会有什么问题？

#### 解析：

```
默认情况下  make 初始化的  channel  是无缓冲的，也就是在迭代写时会阻塞。
```
### 10、以下代码能编译过去吗？为什么？

```
1 type UserAges struct {
```
(^2) ages map[string]int
(^3) sync.Mutex
(^4) }
(^56) func (ua *UserAges) Add(name string, age int) {
7 ua.Lock()
8 defer ua.Unlock()
(^9) ua.ages[name] = age
(^10) }
(^1112) func (ua *UserAges) Get(name string) int {
(^13) if age, ok := ua.ages[name]; ok {
(^14) return age
(^15) }
16 return -
17 }
1 func (set *threadSafeSet) Iter() <-chan interface{} {
(^2) ch := make(chan interface{})
(^3) go func() {
4 set.RLock()
56 for elem := range set.s {
7 ch <- elem
(^8) }
(^109) close(ch)
(^11) set.RUnlock()
(^1213) }()
(^14) return ch
(^15) }


#### 解析：

```
编译失败，值类型  Student{} 未实现接口 People 的方法，不能定义为 People 类
型。
在 golang 语言中， Student  和  *Student  是两种类型，第一个是表示 Student  本
身，第二个是指向  Student  的指针。
```
### 11、以下代码打印出来什么内容，说出为什么。。。

```
1 package main
23 import (
4 "fmt"
5 )
```
(^67) type People interface {
(^8) Speak(string) string
(^9) }
(^1011) type Student struct{}
(^1213) func (stu *Student) Speak(think string) (talk string) {
(^14) if think == "bitch" {
15 talk = "You are a good boy"
16 } else {
17 talk = "hi"
(^18) }
(^19) return
(^20) }
(^2122) func main() {
(^23) var peo People = Student{}
24 think := "bitch"
25 fmt.Println(peo.Speak(think))
26 }
1 package main
(^23) import (
(^4) "fmt"
(^5) )
67 type People interface {
8 Show()
9 }
(^1011) type Student struct{}
(^1213) func (stu *Student) Show() {
(^1415) }
(^1617) func live() People {
(^18) var stu *Student
19 return stu
20 }
2122 func main() {
(^23) if live() == nil {
(^24) fmt.Println("AAAAAAA")


#### 解析：

```
跟上一题一样，不同的是 *Student  的定义后本身没有初始化值，所
以  *Student 是  nil 的，但是 *Student  实现了 People  接口，接口不为  nil 。
```
### 在 golang 协程和channel配合使用

```
写代码实现两个 goroutine，其中一个产生随机数并写入到 go channel 中，另外一
个从 channel 中读取数字并打印到标准输出。最终输出五个随机数。
解析
这是一道很简单的golang基础题目，实现方法也有很多种，一般想答让面试官满意的答
案还是有几点注意的地方。
```
1. goroutine  在golang中式非阻塞的
2. channel  无缓冲情况下，读写都是阻塞的，且可以用 for 循环来读取数据，当管道
    关闭后， for  退出。
3.golang 中有专用的 select case  语法从管道读取数据。
示例代码如下：

### 实现阻塞读且并发安全的map

```
GO里面MAP如何实现key不存在 get操作等待 直到key存在或者超时，保证并发安全，
且需要实现以下接口：
```
25 } else {
26 fmt.Println("BBBBBBB")

(^27) }
(^28) }
1 func main() {
(^2) out := make(chan int)
(^3) wg := sync.WaitGroup{}
(^4) wg.Add(2)
5 go func() {
6 defer wg.Done()
7 for i := 0; i < 5; i++ {
(^8) out <- rand.Intn(5)
(^9) }
(^10) close(out)
(^11) }()
(^12) go func() {
(^13) defer wg.Done()
14 for i := range out {
15 fmt.Println(i)
16 }
(^17) }()
(^18) wg.Wait()
(^19) }
1 type sp interface {


#### 解析：

```
看到阻塞协程第一个想到的就是 channel ，题目中要求并发安全，那么必须用锁，还要
实现多个 goroutine 读的时候如果值不存在则阻塞，直到写入值，那么每个键值需要有
一个阻塞 goroutine  的  channel 。
实现如下：
```
### 高并发下的锁与map的读写

```
场景：在一个高并发的web服务器中，要限制IP的频繁访问。现模拟100个IP同时并发访问服
务器，每个IP要重复访问1000次。
每个IP三分钟之内只能访问一次。修改以下代码完成该过程，要求能成功输出 success:
```
```
Out(key string, val interface{}) //存入key /val，如果该key读取的
goroutine挂起，则唤醒。此方法不会阻塞，时刻都可以立即执行并返回
```
```
2
```
```
Rd(key string, timeout time.Duration) interface{} //读取一个key，如果
key不存在阻塞，等待key存在或者超时
```
```
3
```
(^4) }
1 type Map struct {
(^2) c map[string]*entry
(^3) rmx *sync.RWMutex
(^4) }
(^5) type entry struct {
(^6) ch chan struct{}
(^7) value interface{}
8 isExist bool
9 }
1011 func (m *Map) Out(key string, val interface{}) {
(^12) m.rmx.Lock()
(^13) defer m.rmx.Unlock()
(^14) item, ok := m.c[key]
(^15) if !ok {
(^16) m.c[key] = &entry{
17 value: val,
18 isExist: true,
19 }
(^20) return
(^21) }
(^22) item.value = val
(^23) if !item.isExist {
(^24) if item.ch != nil {
(^25) close(item.ch)
26 item.ch = nil
27 }
28 }
(^29) return
(^30) }
1 package main
(^2)


```
解析
该问题主要考察了并发情况下map的读写问题，而给出的初始代码，又存在 for 循环中启动
goroutine 时变量使用问题以及 goroutine 执行滞后问题。
因此，首先要保证启动的 goroutine 得到的参数是正确的，然后保证 map 的并发读写，最
后保证三分钟只能访问一次。
多CPU核心下修改 int 的值极端情况下会存在不同步情况，因此需要原子性的修改int值。
下面给出的实例代码，是启动了一个协程每分钟检查一下 map 中的过期 ip， for 启动协
程时传参。
```
```
3 import (
4 "fmt"
```
(^5) "time"
(^6) )
(^7)
(^8) type Ban struct {
(^9) visitIPs map[string]time.Time
10 }
11
12 func NewBan() *Ban {
(^13) return &Ban{visitIPs: make(map[string]time.Time)}
(^14) }
(^15) func (o *Ban) visit(ip string) bool {
(^16) if _, ok := o.visitIPs[ip]; ok {
(^17) return true
(^18) }
19 o.visitIPs[ip] = time.Now()
20 return false
21 }
(^22) func main() {
(^23) success := 0
(^24) ban := NewBan()
(^25) for i := 0; i < 1000; i++ {
(^26) for j := 0; j < 100; j++ {
27 go func() {
28 ip := fmt.Sprintf("192.168.1.%d", j)
29 if !ban.visit(ip) {
(^30) success++
(^31) }
(^32) }()
(^33) }
(^34)
(^35) }
36 fmt.Println("success:", success)
37 }
1 package main
23 import (
4 "context"
(^5) "fmt"
(^6) "sync"


```
7 "sync/atomic"
8 "time"
```
(^9) )
(^1011) type Ban struct {
(^12) visitIPs map[string]time.Time
(^13) lock sync.Mutex
(^14) }
1516 func NewBan(ctx context.Context) *Ban {
17 o := &Ban{visitIPs: make(map[string]time.Time)}
18 go func() {
(^19) timer := time.NewTimer(time.Minute * 1)
(^20) for {
(^21) select {
(^22) case <-timer.C:
(^23) o.lock.Lock()
(^24) for k, v := range o.visitIPs {
25 if time.Now().Sub(v) >= time.Minute*1 {
26 delete(o.visitIPs, k)
27 }
(^28) }
(^29) o.lock.Unlock()
(^30) timer.Reset(time.Minute * 1)
(^31) case <-ctx.Done():
(^32) return
33 }
34 }
35 }()
(^36) return o
(^37) }
(^38) func (o *Ban) visit(ip string) bool {
(^39) o.lock.Lock()
(^40) defer o.lock.Unlock()
(^41) if _, ok := o.visitIPs[ip]; ok {
42 return true
43 }
44 o.visitIPs[ip] = time.Now()
(^45) return false
(^46) }
(^47) func main() {
(^48) success := int64(0)
(^49) ctx, cancel := context.WithCancel(context.Background())
50 defer cancel()
5152 ban := NewBan(ctx)
5354 wait := &sync.WaitGroup{}
(^5556) wait.Add(1000 * 100)
(^57) for i := 0; i < 1000; i++ {
(^58) for j := 0; j < 100; j++ {
(^59) go func(j int) {
(^60) defer wait.Done()


### 写出以下逻辑，要求每秒钟调用一次proc并保证程序不退出?

```
解析
题目主要考察了两个知识点：
1.定时执行执行任务
2.捕获 panic 错误
题目中要求每秒钟执行一次，首先想到的就是 time.Ticker 对象，该函数可每秒钟往 chan
中放一个 Time ,正好符合我们的要求。
在  golang  中捕获  panic  一般会用到  recover()  函数。
```
61 ip := fmt.Sprintf("192.168.1.%d", j)
62 if !ban.visit(ip) {

(^63) atomic.AddInt64(&success, 1)
(^64) }
(^65) }(j)
(^66) }
(^6768) }
69 wait.Wait()
7071 fmt.Println("success:", success)
72 }
1 package main
(^23) func main() {
(^4) go func() {
(^5) // 1 在这里需要你写算法
(^6) // 2 要求每秒钟调用一次proc函数
(^7) // 3 要求程序不能退出
(^8) }()
109 select {}
11 }
1213 func proc() {
(^14) panic("ok")
(^15) }
1 package main
(^23) import (
(^4) "fmt"
(^5) "time"
(^6) )
(^78) func main() {
(^9) go func() {
10 // 1 在这里需要你写算法
11 // 2 要求每秒钟调用一次proc函数
12 // 3 要求程序不能退出
(^1314) t := time.NewTicker(time.Second * 1)
(^15) for {
(^16) select {
(^17) case <-t.C:
(^18) go func() {


### 为 sync.WaitGroup 中Wait函数支持 WaitTimeout 功能.

```
解析
```
19 defer func() {
20 if err := recover(); err != nil {

(^21) fmt.Println(err)
(^22) }
(^23) }()
(^24) proc()
(^25) }()
26 }
27 }
28 }()
(^2930) select {}
(^31) }
(^3233) func proc() {
(^34) panic("ok")
(^35) }
1 package main
(^23) import (
(^4) "fmt"
5 "sync"
6 "time"
7 )
(^89) func main() {
(^10) wg := sync.WaitGroup{}
(^11) c := make(chan struct{})
(^12) for i := 0; i < 10; i++ {
(^13) wg.Add(1)
(^14) go func(num int, close <-chan struct{}) {
15 defer wg.Done()
16 <-close
(^17) fmt.Println(num)
(^18) }(i, c)
(^19) }
(^2021) if WaitTimeout(&wg, time.Second*5) {
(^22) close(c)
(^23) fmt.Println("timeout exit")
24 }
25 time.Sleep(time.Second * 10)
26 }
(^2728) func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
(^29) // 要求手写代码
(^30) // 要求sync.WaitGroup支持timeout功能
(^31) // 如果timeout到了超时时间返回true
(^32) // 如果WaitGroup自然结束返回false
(^33) }


```
首先  sync.WaitGroup 对象的  Wait  函数本身是阻塞的，同时，超时用到的
time.Timer 对象也需要阻塞的读。
同时阻塞的两个对象肯定要每个启动一个协程,每个协程去处理一个阻塞，难点在于怎么知道
哪个阻塞先完成。
目前我用的方式是声明一个没有缓冲的chan ，谁先完成谁优先向管道中写入数据。
```
### 语法找错题

```
1 package main
23 import (
```
(^4) "fmt"
(^5) "sync"
(^6) "time"
(^7) )
(^89) func main() {
10 wg := sync.WaitGroup{}
11 c := make(chan struct{})
12 for i := 0; i < 10; i++ {
(^13) wg.Add(1)
(^14) go func(num int, close <-chan struct{}) {
(^15) defer wg.Done()
(^16) <-close
(^17) fmt.Println(num)
(^18) }(i, c)
19 }
2021 if WaitTimeout(&wg, time.Second*5) {
22 close(c)
(^23) fmt.Println("timeout exit")
(^24) }
(^25) time.Sleep(time.Second * 10)
(^26) }
(^2728) func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
29 // 要求手写代码
30 // 要求sync.WaitGroup支持timeout功能
31 // 如果timeout到了超时时间返回true
(^32) // 如果WaitGroup自然结束返回false
(^33) ch := make(chan bool, 1)
(^3435) go time.AfterFunc(timeout, func() {
(^36) ch <- true
(^37) })
(^3839) go func() {
40 wg.Wait()
41 ch <- false
42 }()
(^43)
(^44) return <- ch
(^45) }


### 写出以下代码出现的问题

```
golang 中字符串是不能赋值 nil  的，也不能跟  nil  比较。
```
### 写出以下打印内容

### 找出下面代码的问题

```
1 package main
```
(^2) import (
(^3) "fmt"
4 )
5 func main() {
(^6) var x string = nil
(^7) if x == nil {
(^8) x = "default"
(^9) }
(^10) fmt.Println(x)
(^11) }
1 package main
2 import "fmt"
(^3) const (
(^4) a = iota
(^5) b = iota
(^6) )
(^7) const (
(^8) name = "menglu"
9 c = iota
10 d = iota
11 )
(^12) func main() {
(^13) fmt.Println(a)
(^14) fmt.Println(b)
(^15) fmt.Println(c)
(^16) fmt.Println(d)
17 }
1 package main
2 import "fmt"
3 type query func(string) string
(^45) func exec(name string, vs ...query) string {
(^6) ch := make(chan string)


```
上面的代码有严重的内存泄漏问题，出错的位置是  go fn(i) ，实际上代码执行后会启动
4 个协程，但是因为  ch 是非缓冲的，只可能有一个协程写入成功。而其他三个协程
会一直在后台等待写入。
```
### 写出以下打印结果，并解释下为什么这么打印的。

```
golang 中的切片底层其实使用的是数组。当使用str1[1:]  使， str2 和  str1  底层共
享一个数组，这回导致  str2[1] = "new"  语句影响  str1。
而  append 会导致底层数组扩容，生成新的数组，因此追加数据后的  str2 不会影
响  str1。
但是为什么对 str2  复制后影响的确实 str1  的第三个元素呢？这是因为切
片  str2 是从数组的第二个元素开始， str2 索引为 1 的元素对应的是 str1  索引为
2 的元素。
```
```
7 fn := func(i int) {
8 ch <- vs[i](name)
```
(^9) }
(^10) for i, _ := range vs {
(^11) go fn(i)
(^12) }
(^13) return <-ch
14 }
1516 func main() {
17 ret := exec("111", func(n string) string {
(^18) return n + "func1"
(^19) }, func(n string) string {
(^20) return n + "func2"
(^21) }, func(n string) string {
(^22) return n + "func3"
(^23) }, func(n string) string {
24 return n + "func4"
25 })
26 fmt.Println(ret)
(^27) }
1 package main
2 import (
3 "fmt"
4 )
(^5) func main() {
(^6) str1 := []string{"a", "b", "c"}
(^7) str2 := str1[1:]
(^8) str2[1] = "new"
(^9) fmt.Println(str1)
(^10) str2 = append(str2, "z", "x", "y")
11 fmt.Println(str1)
12 }


### 写出以下打印结果

#### 个人理解：指针类型比较的是指针地址，非指针类型比较的是每个属性的值。

### 写出以下代码的问题

#### 数组只能与相同纬度⻓度以及类型的其他数组比较，切片之间不能直接比较。。

### 下面代码写法有什么问题？

```
1 package main
```
(^23) import (
(^4) "fmt"
5 )
67 type Student struct {
(^8) Name string
(^9) }
(^1011) func main() {
(^12) fmt.Println(&Student{Name: "menglu"} == &Student{Name: "menglu"})
(^13) fmt.Println(Student{Name: "menglu"} == Student{Name: "menglu"})
(^14) }
1 package main
23 import (
(^4) "fmt"
(^5) )
(^67) func main() {
(^8) fmt.Println([...]string{"1"} == [...]string{"1"})
(^9) fmt.Println([]string{"1"} == []string{"1"})
(^10) }
1 package main
2 import (
(^3) "fmt"
(^4) )
(^5) type Student struct {
(^6) Age int
(^7) }
(^8) func main() {
9 kv := map[string]Student{"menglu": {Age: 21}}
10 kv["menglu"].Age = 22
11 s := []Student{{Age: 21}}
(^12) s[0].Age = 22
(^13) fmt.Println(kv, s)


```
golang中的 map  通过 key 获取到的实际上是两个值，第一个是获取到的值，第二个
是是否存在该key 。因此不能直接通过 key 来赋值对象。
```
### golang 并发题目测试

```
题目来源： Go并发编程小测验： 你能答对几道题？
```
### 1 Mutex

####  A: 不能编译

```
 B: 输出 main --> A --> B --> C
 C: 输出 main
 D: panic
```
### 2 RWMutex

14 }

```
1 package main
```
(^2) import (
(^3) "fmt"
(^4) "sync"
(^5) )
(^6) var mu sync.Mutex
(^7) var chain string
8 func main() {
9 chain = "main"
10 A()
(^11) fmt.Println(chain)
(^12) }
(^13) func A() {
(^14) mu.Lock()
(^15) defer mu.Unlock()
16 chain = chain + " --> A"
17 B()
18 }
(^19) func B() {
(^20) chain = chain + " --> B"
(^21) C()
(^22) }
(^23) func C() {
(^24) mu.Lock()
25 defer mu.Unlock()
26 chain = chain + " --> C"
27 }
1 package main


####  A: 不能编译

####  B: 输出 1

```
 C: 程序hang住
 D: panic
```
### 3 Waitgroup

```
2 import (
3 "fmt"
```
(^4) "sync"
(^5) "time"
(^6) )
(^7) var mu sync.RWMutex
(^8) var count int
9 func main() {
10 go A()
11 time.Sleep(2 * time.Second)
(^12) mu.Lock()
(^13) defer mu.Unlock()
(^14) count++
(^15) fmt.Println(count)
(^16) }
(^17) func A() {
18 mu.RLock()
19 defer mu.RUnlock()
20 B()
(^21) }
(^22) func B() {
(^23) time.Sleep(5 * time.Second)
(^24) C()
(^25) }
26 func C() {
27 mu.RLock()
28 defer mu.RUnlock()
(^29) }
1 package main
(^2) import (
(^3) "sync"
(^4) "time"
5 )
6 func main() {
7 var wg sync.WaitGroup
(^8) wg.Add(1)
(^9) go func() {
(^10) time.Sleep(time.Millisecond)
(^11) wg.Done()


####  A: 不能编译

####  B: 无输出，正常退出

```
 C: 程序hang住
 D: panic
```
### 4 双检查实现单例

####  A: 不能编译

####  B: 可以编译，正确实现了单例

```
 C: 可以编译，有并发问题，f函数可能会被执行多次
 D: 可以编译，但是程序运行会panic
```
### 5 Mutex

12 wg.Add(1)
13 }()

(^14) wg.Wait()
(^15) }
1 package doublecheck
(^2) import (
3 "sync"
4 )
5 type Once struct {
(^6) m sync.Mutex
(^7) done uint32
(^8) }
(^9) func (o *Once) Do(f func()) {
(^10) if o.done == 1 {
(^11) return
12 }
13 o.m.Lock()
14 defer o.m.Unlock()
(^15) if o.done == 0 {
(^16) o.done = 1
(^17) f()
(^18) }
(^19) }
1 package main
2 import (
(^3) "fmt"
(^4) "sync"
(^5) )
(^6) type MyMutex struct {
(^7) count int


####  A: 不能编译

####  B: 输出 1, 1

####  C: 输出 1, 2

```
 D: panic
```
### 6 Pool

```
8 sync.Mutex
9 }
```
(^10) func main() {
(^11) var mu MyMutex
(^12) mu.Lock()
(^13) var mu2 = mu
(^14) mu.count++
15 mu.Unlock()
16 mu2.Lock()
17 mu2.count++
(^18) mu2.Unlock()
(^19) fmt.Println(mu.count, mu2.count)
(^20) }
1 package main
(^2) import (
3 "bytes"
4 "fmt"
5 "runtime"
(^6) "sync"
(^7) "time"
(^8) )
(^9) var pool = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
(^10) func main() {
11 go func() {
12 for {
13 processRequest(1 << 28) // 256MiB
(^14) }
(^15) }()
(^16) for i := 0; i < 1000; i++ {
(^17) go func() {
(^18) for {
(^19) processRequest(1 << 10) // 1KiB
20 }
21 }()
22 }
(^23) var stats runtime.MemStats
(^24) for i := 0; ; i++ {
(^25) runtime.ReadMemStats(&stats)
(^26) fmt.Printf("Cycle %d: %dB\n", i, stats.Alloc)


####  A: 不能编译

####  B: 可以编译，运行时正常，内存稳定

####  C: 可以编译，运行时内存可能暴涨

####  D: 可以编译，运行时内存先暴涨，但是过一会会回收掉

### 7 channel

####  A: 不能编译

```
 B: 一段时间后总是输出  #goroutines: 1
 C: 一段时间后总是输出 #goroutines: 2
 D: panic
```
27 time.Sleep(time.Second)
28 runtime.GC()

(^29) }
(^30) }
(^31) func processRequest(size int) {
(^32) b := pool.Get().(*bytes.Buffer)
(^33) time.Sleep(500 * time.Millisecond)
34 b.Grow(size)
35 pool.Put(b)
36 time.Sleep(1 * time.Millisecond)
(^37) }
1 package main
(^2) import (
(^3) "fmt"
(^4) "runtime"
5 "time"
6 )
7 func main() {
(^8) var ch chan int
(^9) go func() {
(^10) ch = make(chan int, 1)
(^11) ch <- 1
(^12) }()
13 go func(ch chan int) {
14 time.Sleep(time.Second)
15 <-ch
(^16) }(ch)
(^17) c := time.Tick(1 * time.Second)
(^18) for range c {
(^19) fmt.Printf("#goroutines: %d\n", runtime.NumGoroutine())
(^20) }
(^21) }


### 8 channel

####  A: 不能编译

####  B: 输出 1

####  C: 输出 0

```
 D: panic
```
### 9 Map

####  A: 不能编译

####  B: 输出 1

####  C: 输出 0

```
 D: panic
```
### 10 happens before

```
1 package main
```
(^2) import "fmt"
(^3) func main() {
(^4) var ch chan int
(^5) var count int
6 go func() {
7 ch <- 1
(^8) }()
(^9) go func() {
(^10) count++
(^11) close(ch)
(^12) }()
(^13) <-ch
14 fmt.Println(count)
15 }
1 package main
(^2) import (
(^3) "fmt"
(^4) "sync"
(^5) )
(^6) func main() {
7 var m sync.Map
8 m.LoadOrStore("a", 1)
(^9) m.Delete("a")
(^10) fmt.Println(m.Len())
(^11) }
1 package main


####  A: 不能编译

####  B: 输出 1

####  C: 输出 0

```
 D: panic
```
### 答案

### 1. D

```
会产生死锁 panic ，因为Mutex  是互斥锁。
```
### 2. D

```
会产生死锁 panic ，根据sync/rwmutex.go  中注释可以知道，读写锁当有一个协程在
等待写锁时，其他协程是不能获得读锁的，而在 A 和C 中同一个调用链中间需要让出
读锁，让写锁优先获取，而 A 的读锁又要求C 调用完成，因此死锁。
```
### 3. D

```
WaitGroup  在调用 Wait  之后是不能再调用 Add  方法的。
```
### 4. C

#### 在多核CPU中，因为CPU缓存会导致多个核心中变量值不同步。

### 5. D

```
加锁后复制变量，会将锁的状态也复制，所以mu1  其实是已经加锁状态，再加锁会死
锁。
```
```
2 var c = make(chan int)
3 var a int
```
(^4) func f() {
(^5) a = 1
(^6) <-c
(^7) }
(^8) func main() {
9 go f()
10 c <- 0
11 print(a)
(^12) }


### 6. C

#### 个人理解，在单核CPU中，内存可能会稳定在 256MB ，如果是多核可能会暴涨。

### 7. C

因为 ch  未初始化，写和读都会阻塞，之后被第一个协程重新赋值，导致写的 ch  都
阻塞。

### 8. D

```
ch  未有被初始化，关闭时会报错。
```
### 9. A

```
sync.Map 没有  Len 方法。
```
### 10. B

```
c <- 0  会阻塞依赖于  f()  的执行。
```
### 记一道字节跳动的算法面试题

### 题目

这其实是一道变形的链表反转题，大致描述如下 给定一个单链表的头节点 head,实现一
个调整单链表的函数，使得每K个节点之间为一组进行逆序，并且从链表的尾部开始组
起，头部剩余节点数量不够一组的不需要逆序。（不能使用队列或者栈作为辅助）
例如：
链表:1->2->3->4->5->6->7->8->null, K = 3 。那么  6->7->8 ，3->4->5 ， 1->2 各
位一组。调整后：1->2->5->4->3->8->7->6->null 。其中 1，2不调整，因为不够一
组。
解析
原文： https://juejin.im/post/5d4f76325188253b49244dd0

### 多协程查询切片问题


### 题目

```
假设有一个超⻓的切片，切片的元素类型为int，切片中的元素为乱序排序。限时5秒，
使用多个goroutine查找切片中是否存在给定的值，在查找到目标值或者超时后立刻结
束所有goroutine的执行。
比如，切片  [23,32,78,43,76,65,345,762,......915,86] ，查找目标值为 345 ，如果切片中
存在，则目标值输出 "Found it!" 并立即取消仍在执行查询任务的 goroutine 。
如果在超时时间未查到目标值程序，则输出 "Timeout！Not Found" ，同时立即取消仍在
执行的查找任务的goroutine 。
答案: https://mp.weixin.qq.com/s/GhC2WDw3VHP91DrrFVCnag
```
### 对已经关闭的的chan进行读写，会怎么样？为什么？

### 题目

```
对已经关闭的的 chan 进行读写，会怎么样？为什么？
```
### 回答

```
 读已经关闭的 chan 能一直读到东⻄，但是读到的内容根据通道内关闭前是否有元素
而不同。
 如果 chan 关闭前，buffer 内有元素还未读 , 会正确读到 chan 内的值，且返回的第二
个 bool 值（是否读成功）为 true。
 如果 chan 关闭前，buffer 内有元素已经被读完，chan 内无值，接下来所有接收的值
都会非阻塞直接成功，返回 channel 元素的零值，但是第二个 bool 值一直为 false。
 写已经关闭的 chan 会 panic
```
### 示例

### 1. 写已经关闭的 chan

```
1 func main(){
2 c := make(chan int,3)
```
(^3) close(c)
(^4) c <- 1
(^5) }
(^6) //输出结果
(^7) panic: send on closed channel
(^89) goroutine 1 [running]
10 main.main()
11 ...


```
 注意这个 send on closed channel，待会会提到。
```
### 2. 读已经关闭的 chan

#### 输出结果

```
1 package main
```
(^2) import "fmt"
(^34) func main() {
(^5) fmt.Println("以下是数值的chan")
6 ci:=make(chan int,3)
7 ci<-1
8 close(ci)
(^9) num,ok := <- ci
(^10) fmt.Printf("读chan的协程结束，num=%v， ok=%v\n",num,ok)
(^11) num1,ok1 := <-ci
(^12) fmt.Printf("再读chan的协程结束，num=%v， ok=%v\n",num1,ok1)
(^13) num2,ok2 := <-ci
(^14) fmt.Printf("再再读chan的协程结束，num=%v， ok=%v\n",num2,ok2)
15
16 fmt.Println("以下是字符串chan")
17 cs := make(chan string,3)
(^18) cs <- "aaa"
(^19) close(cs)
(^20) str,ok := <- cs
(^21) fmt.Printf("读chan的协程结束，str=%v， ok=%v\n",str,ok)
(^22) str1,ok1 := <-cs
23 fmt.Printf("再读chan的协程结束，str=%v， ok=%v\n",str1,ok1)
24 str2,ok2 := <-cs
25 fmt.Printf("再再读chan的协程结束，str=%v， ok=%v\n",str2,ok2)
(^2627) fmt.Println("以下是结构体chan")
(^28) type MyStruct struct{
(^29) Name string
(^30) }
(^31) cstruct := make(chan MyStruct,3)
(^32) cstruct <- MyStruct{Name: "haha"}
33 close(cstruct)
34 stru,ok := <- cstruct
35 fmt.Printf("读chan的协程结束，stru=%v， ok=%v\n",stru,ok)
(^36) stru1,ok1 := <-cs
(^37) fmt.Printf("再读chan的协程结束，stru=%v， ok=%v\n",stru1,ok1)
(^38) stru2,ok2 := <-cs
(^39) fmt.Printf("再再读chan的协程结束，stru=%v， ok=%v\n",stru2,ok2)
(^40) }
1 以下是数值的chan
2 读chan的协程结束，num=1， ok=true
3 再读chan的协程结束，num=0， ok=false


### 多问一句

### 1. 为什么写已经关闭的  chan  就会  panic  呢？

```
 当  c.closed != 0  则为通道关闭，此时执行写，源码提示直接  panic ，输出的内容
就是上面提到的 "send on closed channel" 。
```
### 2. 为什么读已关闭的 chan 会一直能读到值？

```
4 再再读chan的协程结束，num=0， ok=false
5 以下是字符串chan
```
(^6) 读chan的协程结束，str=aaa， ok=true
(^7) 再读chan的协程结束，str=， ok=false
(^8) 再再读chan的协程结束，str=， ok=false
(^9) 以下是结构体chan
(^10) 读chan的协程结束，stru={haha}， ok=true
11 再读chan的协程结束，stru=， ok=false
12 再再读chan的协程结束，stru=， ok=false
1 //在 src/runtime/chan.go
func chansend(c *hchan,ep unsafe.Pointer,block bool,callerpc uintptr) bool
{
2
(^3) //省略其他
(^4) if c.closed != 0 {
(^5) unlock(&c.lock)
(^6) panic(plainError("send on closed channel"))
(^7) }
(^8) //省略其他
9 }
func chanrecv(c *hchan,ep unsafe.Pointer,block bool) (selected,received
bool) {
1
2 //省略部分逻辑
(^3) lock(&c.lock)
(^4) //当chan被关闭了，而且缓存为空时
(^5) //ep 是指 val,ok := <-c 里的val地址
(^6) if c.closed != 0 && c.qcount == 0 {
(^7) if receenabled {
(^8) raceacquire(c.raceaddr())
9 }
10 unlock(&c.lock)
11 //如果接受之的地址不空，那接收值将获得一个该值类型的零值
(^12) //typedmemclr 会根据类型清理响应的内存
(^13) //这就解释了上面代码为什么关闭的chan 会返回对应类型的零值


```
 c.closed != 0 && c.qcount == 0  指通道已经关闭，且缓存为空的情况下（已经读完
了之前写到通道里的值）
 如果接收值的地址  ep 不为空
 那接收值将获得是一个该类型的零值
 typedmemclr  会根据类型清理相应地址的内存
 这就解释了上面代码为什么关闭的 chan 会返回对应类型的零值
```
### 简单聊聊内存逃逸？

### 问题

```
知道golang的内存逃逸吗？什么情况下会发生内存逃逸？
```
### 回答

```
golang程序变量会携带有一组校验数据，用来证明它的整个生命周期是否在运行时完全
可知。如果变量通过了这些校验，它就可以在栈上分配。否则就说它 逃逸 了，必须在
堆上分配。
能引起变量逃逸到堆上的典型情况：
 在方法内把局部变量指针返回 局部变量原本应该在栈中分配，在栈中回收。但是由
于返回时被外部引用，因此其生命周期大于栈，则溢出。
 发送指针或带有指针的值到 channel 中。 在编译时，是没有办法知道哪
个  goroutine  会在 channel  上接收数据。所以编译器没法知道变量什么时候才会
被释放。
 在一个切片上存储指针或带指针的值。 一个典型的例子就是  []*string 。这会导致
切片的内容逃逸。尽管其后面的数组可能是在栈上分配的，但其引用的值一定是在
堆上。
 slice 的背后数组被重新分配了，因为 append 时可能会超出其容量( cap )。 slice
初始化的地方在编译时是可以知道的，它最开始会在栈上分配。如果切片背后的存
储要基于运行时的数据进行扩充，就会在堆上分配。
 在 interface 类型上调用方法。 在 interface 类型上调用方法都是动态调度的 ——
方法的真正实现只能在运行时知道。想像一个 io.Reader 类型的变量 r , 调用
r.Read(b) 会使得 r 的值和切片b 的背后存储都逃逸掉，所以会在堆上分配。
```
14 if ep != null {
15 typedmemclr(c.elemtype,ep)

(^16) }
(^17) //返回两个参数 selected,received
(^18) // 第二个采纳数就是 val,ok := <- c 里的 ok
(^19) //也就解释了为什么读关闭的chan会一直返回false
(^20) return true,false
21 }
22 }


### 举例

```
通过一个例子加深理解，接下来尝试下怎么通过 go build -gcflags=-m  查看逃逸的情
况。
```
```
执行go build -gcflags=-m main.go
```
```
 ./main.go:8:10: new(A) escapes to heap  说明  new(A) 逃逸了,符合上述提到的常⻅
情况中的第一种。
 ./main.go:14:11: main a.s + " world" does not escape  说明 b 变量没有逃逸，因为它
只在方法内存在，会在方法结束时被回收。
 ./main.go:15:9: b + "!" escapes to heap  说明 c 变量逃逸，通过 fmt.Println(a
...interface{}) 打印的变量，都会发生逃逸，感兴趣的朋友可以去查查为什么。
```
```
1 package main
2 import "fmt"
3 type A struct {
```
(^4) s string
(^5) }
(^6) // 这是上面提到的 "在方法内把局部变量指针返回" 的情况
(^7) func foo(s string) *A {
(^8) a := new(A)
(^9) a.s = s
10 return a //返回局部变量a,在C语言中妥妥野指针，但在go则ok，但a会逃逸到堆
11 }
12 func main() {
(^13) a := foo("hello")
(^14) b := a.s + " world"
(^15) c := b + "!"
(^16) fmt.Println(c)
(^17) }
1 go build -gcflags=-m main.go
2 # command-line-arguments
3 ./main.go:7:6: can inline foo
4 ./main.go:13:10: inlining call to foo
(^5) ./main.go:16:13: inlining call to fmt.Println
/var/folders/45/qx9lfw2s2zzgvhzg3mtzkwzc0000gn/T/go-
build409982591/b001/_gomod_.go:6:6: can inline init.0
6
(^7) ./main.go:7:10: leaking param: s
(^8) ./main.go:8:10: new(A) escapes to heap
9 ./main.go:16:13: io.Writer(os.Stdout) escapes to heap
10 ./main.go:16:13: c escapes to heap
11 ./main.go:15:9: b + "!" escapes to heap
(^12) ./main.go:13:10: main new(A) does not escape
(^13) ./main.go:14:11: main a.s + " world" does not escape
(^14) ./main.go:16:13: main []interface {} literal does not escape
(^15) <autogenerated>:1: os.(*File).close .this does not escape


```
以上操作其实就叫逃逸分析。下篇文章，跟大家聊聊怎么用一个比较trick的方法使变量
不逃逸。方便大家在面试官面前秀一波。
原文 https://mp.weixin.qq.com/s/4YYR1eYFIFsNOaTxL4Q-eQ
```
### 字符串转成byte数组，会发生内存拷⻉吗？

### 问题

```
字符串转成byte数组，会发生内存拷⻉吗？
```
### 回答

#### 字符串转成切片，会产生拷⻉。严格来说，只要是发生类型强转都会发生内存拷⻉。那

#### 么问题来了。

#### 频繁的内存拷⻉操作听起来对性能不大友好。有没有什么办法可以在字符串转成切片的

#### 时候不用发生拷⻉呢？

### 解释

```
StringHeader 是字符串在go的底层结构。
```
```
SliceHeader 是切片在go的底层结构。
```
```
1 package main
```
(^23) import (
(^4) "fmt"
(^5) "reflect"
(^6) "unsafe"
7 )
89 func main() {
10 a :="aaa"
(^11) ssh := *(*reflect.StringHeader)(unsafe.Pointer(&a))
(^12) b := *(*[]byte)(unsafe.Pointer(&ssh))
(^13) fmt.Printf("%v",b)
(^14) }
1 type StringHeader struct {
(^2) Data uintptr
3 Len int
4 }
1 type SliceHeader struct {
2 Data uintptr
(^3) Len int
(^4) Cap int


```
那么如果想要在底层转换二者，只需要把 StringHeader 的地址强转成 SliceHeader 就
行。那么go有个很强的包叫 unsafe 。
```
1. unsafe.Pointer(&a) 方法可以得到变量a的地址。
2. (*reflect.StringHeader)(unsafe.Pointer(&a)) 可以把字符串a转成底层结构的形式。
3. (*[]byte)(unsafe.Pointer(&ssh)) 可以把ssh底层结构体转成byte的切片的指针。
4.再通过  * 转为指针指向的实际内容。

## Golang 理论

### Goroutine调度策略

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

