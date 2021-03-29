### 课程大纲

1. "Hello world!"
2. 文件名、关键字、标识符和注释
3. Go程序的一般结构
4. 包的导入
5. 可见性规则

### 上一课作业答案解析

无

### 本次课堂内容

#### 1、Hello world

```
程序1：
// go两种执行方式的区别？
// 1.如果我们先编译生成了可执行文件[二进制程序]，这个文件可以处处运行。`go build main.go`
// 2. go run 方式需要安装go 的 sdk
// 编译的可执行文件会大一些


// 编译指定输出名称为output
// go build -o output


package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello world!你好，世界！")
}


程序2：
// 包的概念
// 每个 Go 文件都属于且仅属于一个包。一个包可以由许多以 .go 为扩展名的源文件组成，因此文件名和包名一般来说都是不相同的,类似其他语言的命名空间或者类库的概念
// 你必须在源文件中非注释的第一行指明这个文件属于哪个包，如：package main。package main 表示一个可独立执行的程序，每个 Go 应用程序都包含一个名为 main 的包。
// 包名都应该使用小写字母
// main 函数是每一个可执行程序所必须包含的，一般来说都是在启动后第一个执行的函数（如果有 init () 函数则会先执行该函数）
package main

import (
	"fmt"
)

func init () {
	fmt.Println("hello 小T")
}

func main () {
	fmt.Println("hello tiner")
}


程序3
// Go编程的注意点
// 1.文件以go扩展名结尾 2.应用程序的执行入口是main()函数 3.Go语言严格区分大小写 4.Go每条语句没有分号 5、每个源文件开始都用package声明，指定其属于哪个包
// 5.Go是逐行编译的 一行只能写一条语句 6、Go不需要在语句后面或者证明后面添加分号结尾【除非多个声明和语句在同一行】 
// 7.定义的变量或者import没有用到，Go编译不会通过。8.大括号成对出现 缺一不可
package main

import (
	"fmt"
)

func main () {
	var x int; var y int
	x,y=1,2
	fmt.Println(x)
	fmt.Println(y)
}
```

#### 2、文件名、关键字、标识符和注释

##### 名称

Go 中函数、变量、常量、类型、语句标签、包名【下面统称为名称】遵循一个简单的规则：以字母或者下划线开头，由数字字母下划线组成。

下面是25个关键字，不能作为名称使用

```
break	default	func	interface	select
case	defer	go	map	struct
chan	else	goto	package	switch
const	fallthrough	if	range	type
continue	for	import	return	var
```

还有36个预定义字符：

```
append	bool	byte	cap	close	complex	complex64	complex128	uint16
copy	false	float32	float64	imag	int	int8	int16	uint32
int32	int64	iota	len	make	new	nil	panic	uint64
print	println	real	recover	string	true	uint	uint8	uintptr
```

在声明中可以使用它们，但非常不建议

注意：

1. 名称本身没有长度限制，但是习惯上Go的编程风格倾向于短名称，特别是作用域比较小的局部变量
2. 名称的作用域越大，就使用越长且更有意义的名称
3. 遇到单词组合名称时，Go程序员更偏向使用“驼峰式”风格。更喜欢大小写字母而不是下划线

##### 注释

**单行注释**是**最常见**的注释形式，你可以在任何地方使用以 `//` 开头的单行注释。多行注释也叫块注释，均已以 `/*` 开头，并以`*/`结尾，且不可以嵌套使用，**多行注释**一般用于**包的文档描述**或注释**成块的代码片段**。

注意：

1. 包都应该有注释，在 `package` 语句之前的块注释将被默认认为是这个包的文档说明，对整体功能做简要的介绍
2. **全局作用域**的类型、常量、变量、函数和被导出的对象都应该有一个合理的注释

### 练习题

### 补充说明

### 课程链接