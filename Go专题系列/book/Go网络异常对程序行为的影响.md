# Go网络异常对程序行为的影响

Go编写网络程序非常的高效，而且有是那么的简单，寥寥几行代码就可以写一个ECHO协议的程序，所以现在很多网络程序都采用Go语言开发。但是网络状况是复杂的，会有很多的异常状况，如果不能很好和正确的处理这些异常状况，会导致网络程序出现莫名其妙的现象，或者hang住。

本文尝试探讨几种网络异常的情况，研究在这些情况下客户端和服务端的的行为，包括连接断掉的检测能力、half-close情况下两端的读写能力、丢包的情况等等。

这是我首次采用微课的方式分享技术内容，本文是视频内容的整理版。 本来是想录制一个10分钟的视频，一不小心录制了半小时。

https://www.bilibili.com/video/av80946416/?zw



### TCP 协议介绍

[![img](D:\www\better_study_for_golang\每日一题\images\tcp1.png)](https://colobu.com/2019/12/28/go-tcp-exceptions/tcp1.png)

tcp的数据格式包含header和payload, header中会包含消息的状态，比如我们常见的`SYN`、`ACK`、`PSH`、`FIN`等。通过 tcpdump可以根据消息的状态进行筛选。

#### 握手

客户端和服务器端建立连接的时候，需要三路握手。

[![img](D:\www\better_study_for_golang\每日一题\images\3wayhandshake.png)](https://colobu.com/2019/12/28/go-tcp-exceptions/3wayhandshake.png)

因为双方都需要和对方同步seq号，所以需要来回确认。服务器把SYN和ACK合并成一条消息，所以最终只需要三次交流就可以了。当然如果你想把SYN和ACK拆开成两个消息也可以，只不过协议栈一般不这样实现。

比如你参加一次相亲聚会，看到一个漂亮的姑娘，你想去搭讪，首先得先了解一下。

```
你： 
姑娘您好，贵庚啊？
姑娘：小女子18，请问大哥您贵庚啊？
你：我81了......
```

这样寒暄之后你们双方就可以进一步的深入的交流了。

#### 分手

[![img](https://colobu.com/2019/12/28/go-tcp-exceptions/4wayhandshake.png)](https://colobu.com/2019/12/28/go-tcp-exceptions/4wayhandshake.png)

客户端和服务器端都可以主动关闭连接。主动关闭的一方我们称之为发起者，被动关闭接收的那一方我们称之为接受者。

发起者要关闭连接，需要发送`FIN`,然后接收者发送`ACK`。这个时候被动者有可能恋恋不舍，还有数据想发送给你，所以接受者这一端它的连接还没有释放，直到它发送`FIN`，发起者回复`ACK`，接收端的连接才释放。

```
姑娘：我要走了
你：再见...
...你依依不舍
你：我也要走了
姑娘：再见
```

### tcpdump

tcpdump是分析网络情况的神器，经常用来分析疑难杂症，并且让狡辩者哑口无言。

[![img](https://colobu.com/2019/12/28/go-tcp-exceptions/tcpdump.png)](https://colobu.com/2019/12/28/go-tcp-exceptions/tcpdump.png)

打印一张tcpdump的小抄放在案头是明智之举。

### 网络异常状况

视频中，我测试了以下6种网络异常情况下的程序响应情况。

[![img](D:\www\better_study_for_golang\每日一题\images\cases.png)](https://colobu.com/2019/12/28/go-tcp-exceptions/cases.png)

使用的代码基本上是从下面的代码修改而来。

server.go

```
package main
import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)
var (
	addr = flag.String("addr", ":8972", "listened address")
)
func main() {
	flag.Parse()
	ln, err := net.Listen("tcp", *addr)
	panicOnErr(err)
	// 接收一个连接
	conn, err := ln.Accept()
	panicOnErr(err)
	clientAddr := conn.RemoteAddr().String()
	// 读 goroutine
	go func() {
		var buf = make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				log.Printf("read err from client %s: %v", clientAddr, err)
				return
			}
			log.Printf("read %d bytes from client %s", n, clientAddr)
		}
	}()
	// 写
	id := 0
	write := func() {
		msg := fmt.Sprintf("sent id: %d from server", id)
		id++
		n, err := conn.Write([]byte(msg))
		if err != nil {
			log.Printf("write err to client %s: %v", clientAddr, err)
			return
		}
		log.Printf("write %d bytes to client %s", n, clientAddr)
	}
	// 继续监听新的连接
	go func() {
		for {
			_, err := ln.Accept()
			if err != nil {
				log.Printf("accept err : %v", err)
			}
		}
	}()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmd := scanner.Text()
		switch cmd {
		case "close_conn":
			conn.Close()
		case "close_ln":
			ln.Close()
		case "write":
			write()
		case "exit", "quit":
			return
		}
	}
}
func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
```

client.go

```
package main
import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)
var (
	addr = flag.String("addr", "127.0.0.1:8972", "server address")
)
func main() {
	flag.Parse()
	// 连接服务器
	conn, err := net.Dial("tcp", *addr)
	panicOnErr(err)
	// 读 goroutine
	go func() {
		var buf = make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				log.Printf("read err from server %s: %v", *addr, err)
				return
			}
			log.Printf("read %d bytes from server %s", n, *addr)
		}
	}()
	// 写
	id := 0
	write := func() {
		msg := fmt.Sprintf("sent clientid: %d from client", id)
		id++
		n, err := conn.Write([]byte(msg))
		if err != nil {
			log.Printf("write err to server %s: %v", *addr, err)
			return
		}
		log.Printf("write %d bytes to server %s", n, *addr)
	}
	// 阻塞在这里避免客户端退出
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmd := scanner.Text()
		switch cmd {
		case "close_conn":
			conn.Close()
		case "write":
			write()
		case "exit", "quit":
			return
		}
	}
}
func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
```



#### 服务器主动关闭连接， 客户端不关闭连接

- 服务器的 `conn.Read`会怎样？
- 服务器继续 `conn.Write`会怎么样?
- 客户端的 `conn.Read`会怎样？
- 客户端继续 `conn.Write`会怎么样?

#### 服务器主动关闭连接， 客户端检测到异常后也关闭连接

- 服务器的 `conn.Read`会怎样？
- 服务器继续 `conn.Write`会怎么样
- 客户端的 `conn.Read`会怎样？
- 客户端继续 `conn.Write`会怎么样

#### 服务器只关闭`Read`

- 服务器的 `conn.Read`会怎样？
- 服务器继续 `conn.Write`会怎么样
- 客户端的 `conn.Read`会怎样？
- 客户端继续 `conn.Write`会怎么样

#### 服务器只关闭`Write`

- 服务器的 `conn.Read`会怎样？
- 服务器继续 `conn.Write`会怎么样
- 客户端的 `conn.Read`会怎样？
- 客户端继续 `conn.Write`会怎么样

#### 服务器被kill掉

- 服务器的 `conn.Read`会怎样？
- 服务器继续 `conn.Write`会怎么样
- 客户端的 `conn.Read`会怎样？
- 客户端继续 `conn.Write`会怎么样

#### 把网线、挖光纤、雷暴机房、防火墙始乱终弃

只分析其中一种情况: `包丢了`

- 服务器的 `conn.Read`会怎样？
- 服务器继续 `conn.Write`会怎么样
- 客户端的 `conn.Read`会怎样？
- 客户端继续 `conn.Write`会怎么样