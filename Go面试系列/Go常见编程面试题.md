# Go常见编程面试题

### （1）交替打印数字和字母

#### 问题描述

使用两个 goroutine  交替打印序列，一个  goroutine  打印数字， 另外一
个  goroutine  打印字母， 最终效果如下：

`1 12AB34CD56EF78GH910IJ1112KL1314MN1516OP1718QR1920ST2122UV2324WX2526YZ`

#### 解题思路

问题很简单，使用 channel 来控制打印的进度。使用两个 channel ，来分别控制数字和
字母的打印序列， 数字打印完成后通过 channel 通知字母打印, 字母打印完成后通知数
字打印，然后周而复始的工作。

#### 源码参考：

```
package main

import (
	"fmt"
)

func main() {
	str := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	num := 1
	letter, number, done := make(chan bool), make(chan bool), make(chan bool)
	go func() {
		for {
			select {
			case <-number:
				fmt.Print(num)
				num++
				fmt.Print(num)
				num++
				letter <- true
			}
		}
	}()
	go func() {
		i := 0
		for {
			select {
			case <-letter:
				for j := 0; j < 2; j++ {
					if i > len(str)-1 {
						done <- true
						return
					}
					fmt.Print(str[i : i+1])
					i++
				}
				number <- true
			}
		}
	}()
	number<-true
	<-done
}
```

#### 源码解析

```
1、这里用到了两个 channel负责通知，letter负责通知打印字母的goroutine来打印字母，
2、number用来通知打印数字的goroutine打印数字。
3、<-done用来等待字母打印完成后退出循环。
```
### （2）判断字符串中字符是否全都不同

#### 问题描述

请实现一个算法，确定一个字符串的所有字符【是否全都不同】。这里我们要求【**不允许使用额外的存储结构**】。 给定一个string，请返回一个bool值,true代表所有字符全都不同，false代表存在相同的字符。 保证字符串中的字符为【ASCII字符】。字符串的的长度小于等于【3000】。

#### 解题思路

这里有几个重点：

- 第一个是 ASCII字符 ， ASCII字符 字符一共有256个，其中128个是常
  用字符，可以在键盘上输入。128之后的是键盘上无法找到的。
- 然后是**全部不同**，也就是字符串中的字符没有重复的，再次，不准使用额外的储存结
  构，且字符串小于等于3000。
- 如果允许其他额外储存结构，这个题目很好做。如果不允许的话，可以使用golang内置
  的方式实现。

#### 源码参考

```
/*
1、golang内置方法 strings.Count ,可以用来判断在一个字符串中包含的另外一个字符串的数量。
2、golang内置方法 strings.Index 和strings.LastIndex ，用来判断指定字符串在另外一个字符串的索引位置，分别是第一次发现位置和最后发现位置。
*/
package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(isUnique("sdfsafs"))
	fmt.Println(isUnique("abcd"))
}

func isUnique(str string) bool {
	// 判断字符串数量
	if strings.Count(str, "") > 3000 {
		return false
	}
	for k, v := range str {
		if v > 127 {
			return false
		}
		// 在其他不同的key位置也查到该字符串即为有相同字符
		if strings.LastIndex(str, str[k:k+1]) != k {
			return false
		}
	}
	return true
}
```

### （3）翻转字符串

#### 问题描述

请实现一个算法，在不使用【额外数据结构和储存空间】的情况下，翻转一个给定的字符串(可以使用单个过程变量)。

给定一个string，请返回一个string，为翻转后的字符串。保证字符串的的长度小于等于5000。

#### 解题思路

翻转字符串其实是将一个字符串以中间字符为轴，前后翻转，即将str[len]赋值给str[0],将str[0] 赋值 str[len]。

#### 源码参考

```
package main

import (
	"fmt"
)

func main() {
	fmt.Println(reverseStr("12345"))
	fmt.Println(reverseStr("abcd"))
}

func reverseStr(s string) (string, bool) {
	l := len(s)
	if l > 5000 {
		return s, false
	}
	sb := []byte(s)
	// 以字符串⻓度的1/2为轴，前后赋值
	for i := 0; i < l/2; i++ {
		sb[i], sb[l-i-1] = sb[l-i-1], sb[i]
	}
	return string(sb), true
}

```
### （4）判断两个给定的字符串排序后是否一致

#### 问题描述

给定两个字符串，请编写程序，确定其中一个字符串的字符重新排列后，能否变成另一个字符串。 这里规定【大小写为不同字符】，且考虑字符串重点空格。给定一个string s1和一个string s2，请返回一个bool，代表两串是否重新排列后可相同。 保证两串的的长度都小于等于5000。

#### 解题思路

首先要保证字符串的长度小于5000。之后只需要一次循环遍历s1中的字符在s2是否都存在即可。

#### 源码参考

```
package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(isRegroup("12345", "45123"))
	fmt.Println(isRegroup("abcd", "aecd"))
}

func isRegroup(s1, s2 string) bool {
	l1 := len(s1)
	l2 := len(s2)
	if l1 > 5000 || l2 > 5000 || l1 != l2 {
		return false
	}
	for i := 0; i < l1; i++ {
		if strings.Count(s1, s1[i:i+1]) != strings.Count(s2, s1[i:i+1]) {
			return false
		}
	}
	return true
}
```
源码解析

```
这里还是使用golang内置方法  strings.Count 来判断字符是否一致。
```
### （5）字符串替换问题

#### 问题描述

请编写一个方法，将字符串中的空格全部替换为“%20”。 假定该字符串有足够的空间存放新增的字符，并且知道字符串的真实的长度(小于等于1000)，同时保证字符串由【大小写的英文字母组成】。 给定一个string为原始的串，返回替换后的string。

#### 解题思路

两个问题，第一个是只能是英文字母，第二个是替换空格。

#### 源码参考

```
package main

import (
	"fmt"
	"strings"
	"unicode"
)

func main() {
	fmt.Println(replaceBlank("45 123"))
	fmt.Println(replaceBlank("ae c d"))
}

func replaceBlank(s string) (string, bool) {
	if len([]rune(s)) > 1000 {
		return s, false
	}
	for _, v := range s {
		if string(v) != " " && unicode.IsLetter(v) == false {
			return s, false
		}
	}
	return strings.Replace(s, " ", "%20", -1), true
}
```

#### 源码解析

```
这里使用了golang内置方法 unicode.IsLetter判断字符是否是字母，之后使用
strings.Replace来替换空格。
```
### （6）机器人坐标问题

#### 问题描述

有一个机器人，给一串指令，L左转 R右转，F前进一步，B后退一步，问最后机器人的坐标，最开始，机器人位于 0 0，方向为正Y。 可以输入重复指令n ： 比如 R2(LF) 这个等于指令 RLFLF。 问最后机器人的坐标是多少？

#### 解题思路

这里的一个难点是解析重复指令。主要指令解析成功，计算坐标就简单了。

#### 源码参考

```
package main

import (
	"fmt"
	"unicode"
)

const (
	Left = iota //0
	Top         // 1
	Right
	Bottom
)

func main() {
	fmt.Println(moves("R2(LF)", 0, 0, Top))
}
func moves(cmd string, x0 int, y0 int, z0 int) (x, y, z int) {
	x, y, z = x0, y0, z0
	repeat := 0
	repeatCmd := ""
	for _, s := range cmd {
		switch {
		case unicode.IsNumber(s):
			repeat = repeat*10 + (int(s) - '0')
		case s == ')':
			for i := 0; i < repeat; i++ {
				x, y, z = moves(repeatCmd, x, y, z)
			}
			repeat = 0
			repeatCmd = ""
		case repeat > 0 && s != '(' && s != ')':
			repeatCmd = repeatCmd + string(s)
		case s == 'L':
			z = (z + 3) % 4
		case s == 'R':
			z = (z + 1) % 4
		case s == 'F':
			switch {
			case z == Left || z == Right:
				x = x + z - 1
			case z == Top || z == Bottom:
				y = y - z + 2
			}
		case s == 'B':
			switch {
			case z == Left || z == Right:
				x = x - z + 1
			case z == Top || z == Bottom:
				y = y + z - 2
			}
		}
	}
	return
}
```

#### 源码解析

这里使用三个值表示机器人当前的状况，分别是：x表示x坐标，y表示y坐标，z表示当
前方向。 L、R 命令会改变值z，F、B命令会改变值x、y。 值x、y的改变还受当前的z值
影响。
如果是重复指令，那么将重复次数和重复的指令存起来递归调用即可。



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

看到阻塞协程第一个想到的就是 channel ，题目中要求并发安全，那么必须用锁，还要
实现多个 goroutine 读的时候如果值不存在则阻塞，直到写入值，那么每个键值需要有
一个阻塞 goroutine  的  channel 。
实现如下：

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

首先  sync.WaitGroup 对象的  Wait  函数本身是阻塞的，同时，超时用到的
time.Timer 对象也需要阻塞的读。
同时阻塞的两个对象肯定要每个启动一个协程,每个协程去处理一个阻塞，难点在于怎么知道
哪个阻塞先完成。
目前我用的方式是声明一个没有缓冲的chan ，谁先完成谁优先向管道中写入数据。