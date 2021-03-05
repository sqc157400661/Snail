# 线上`read: connection reset by peer`问题排查



## 问题描述

线上在用户访问激增的情况下，出现``read: connection reset by peer``错误



访问链路如下

![preview](D:\www\Snail\Go实战系列\6666666.jpg)

go client客户端配置：

```
var httpClient = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   1000 * time.Millisecond,
			KeepAlive: 300 * time.Second,
		}).DialContext,
		MaxIdleConns:        MaxIdleConns,
		MaxIdleConnsPerHost: MaxIdleConnsPerHost,
		IdleConnTimeout:     time.Duration(IdleConnTimeout) * time.Second,
	},
	Timeout: 1000 * time.Millisecond,
}
```

- 重点讲下`transport.IdleConnTimeout`的含义，它的意思是单个长连接，如果`60s`内没有被使用，就不再使用这个长连接发送请求。

Nginx长连接的一些设置参数：

```
http{
    upstream backend{
        keepalive 65;
        #Nginx 1.15.3之后可以设置，这里是默认配置
        keepalive_timeout 65s;
        #Nginx 1.15.3之后可以设置，这里是默认配置
        keepalive_requests 100;
        ...
    }
    server{
        keepalive_timeout 65s;
        keepalive_requests 100;
        ...
    }
}

```

- `server`中的`keepalive_timeout`, 意思是`Nginx`作为服务端，对于客户端的长连接请求，如果20s内，没有收到新的请求，就会关闭这个连接。
- `server`中的`keepalive_requests`，意思是对于客户端的单个长连接，最多处理`100`个请求，处理完`100`个请求后，就关闭这个连接，不再接收新的请求。
- `upstream`中的`keepalive`，意思是这个`upstream`最多的空闲长连接数
- `upstream`中的`keepalive_timeout`，意思是`Nginx`作为客户端，与`upstream`建立长连接后，如果`60s`内没有使用，就关闭这个长连接，不再使用。
- `upstream`中的`keepalive_requests`，意思是`Nginx`作为客户端，与`upstream`建立的长连接，单个连接最多发送`100`个请求，超过之后，就关闭连接。



## 分析问题



## 重现问题



## 知识点



## 参数调优



## 扩展工具

### (1) liunx下建立socket连接

有一个特殊的文件`/dev/tcp`,打开这个文件就类似于发出了一个`socket`调用，建立一个`socket`连接，读写这个文件就相当于在这个`socket`连接中传输数据。

```
/*
	指定服务器名为：127.0.0.1,
	端口号为：80,
	指定文件描述符为8
	<> 表示输入 输出
*/
exec 8<> /dev/tcp/127.0.0.1/8001

/*
	1表示标准输出
	>& 8表示输出重定向到，文件描述符8上（不加& 会认为是简单的输出到文件8）
	"GET / HTTP/1.0\n" 表示用的什么协议和方法 （应用层协议，如http）
*/
echo -e "GET / HTTP/1.0\n" 1>& 8
 
/*
 0表示标准输入
 <& 8表示接受文件描述符8的输入
*/
cat 0<& 8
```

### （2）nestat 命令

```
netstat -s

netstat -natp 查看socket连接（IP+ 端口映射 每个socket是独立隔离的）
```

### （3）tcpdump

```
tcpdump -nn -i eth0 port 8001
```



## 1、先检查go http client 和 服务器keepalive时间

go http client  keepalive =》300s

nginx服务器 =》65s



做好调整



## 2、问题再次出现







3、知识点：







## 4、linux下修改内核参数进行Tcp性能调优

## 1. net.core.netdev_max_backlog

`net.core.netdev_max_backlog`参数表示网卡接受数据包的队列最大长度，在阿里云服务器上，默认值是1000，可以适当调整。

## 2. net.core.somaxconn

`net.core.somaxconn`参数决定了端口监听队列的最大长度，存放的是已经处于ESTABLISHED而没有被用户程序（nginx）接管的TCP连接，默认是128，对于高并发的，或者瞬发大量连接，必须调高该值，否则会直接丢弃连接。

## 3. net.ipv4.tcp_max_orphans

`net.ipv4.tcp_max_orphans`参数决定孤立连接的最大数量。阿里云服务器默认16384，个人感觉没啥鸟用。

## 4. net.ipv4.tcp_max_syn_backlog

`net.ipv4.tcp_max_syn_backlog`参数决定已经收到syn包，但是还没有来得及确认的连接队列，这是传输层的队列，在高并发的情况下，必须调整该值，提高承载能力。

## 5. net.ipv4.tcp_synack_retries

`net.ipv4.tcp_synack_retries`参数决定了发送`SYN+ACK`确认包重试的次数（数量），默认是5，可以调整为2或者3，使其快速失败。

## 6. net.ipv4.tcp_syn_retries

`net.ipv4.tcp_syn_retries`参数，作为客户端，主动建立连接时发送syn包重试的次数，默认6次，可以调整为2次或者三次，快速失败。

## 7. net.ipv4.tcp_abort_on_overflow

`net.ipv4.tcp_abort_on_overflow`参数，当TCP连接已经建立，并塞到程序监听backlog队列时，如果检测到backlog队列已经满员后，TCP连接状态会回退到`SYN+ACK`状态，假装TCP三次握手第三次客户单的`ACK`包没收到，让客户端重传`ACK`，以便快速进入`ESTABLISHED`状态。如果设置了`net.ipv4.tcp_abort_on_overflow`参数，那么在检测到监听backlog 队列已满时，直接发 RST 包给客户端终止此连接，此时客户端程序会收到`104 Connection reset by peer`错误。这个参数很暴力，慎用。参考[这里](http://blog.csdn.net/rain_qingtian/article/details/41864589)

## 8. net.ipv4.tcp_syncookies

`net.ipv4.tcp_syncookies`参数，在TCP三次握手过程中，当服务端收到最初的`SYN`请求时，会检查应用程序的`syn_backlog`队列是否已满。若已满，通常行为是丢弃此`SYN`包。若未满，会再检查应用程序的监听`backlog`队列是否已满。若已满并且系统根据历史记录判断该应用程序不会较快消耗连接时，则丢弃此 SYN 包。如果启用`tcp_syncookies`则在检查到`syn_backlog`队列已满时，不丢弃该`SYN`包，而改用`syncookie`技术进行三次握手。参考[这里](http://blog.csdn.net/rain_qingtian/article/details/41864589)

## 9. net.ipv4.ip_local_port_range

`net.ipv4.ip_local_port_range`参数决定了作为客户端，发起连接时可用的端口范围，对于nginx来说，后抛请求是就是客户端行为，所以高并发场景下也有一定的必要。

## 10. net.ipv4.tcp_tw_reuse

`net.ipv4.tcp_tw_reuse`参数可以重用`TIME_WAIT`状态的连接，仅需要1秒就可以重用。此参数针对`TIME_WAIT`，与是否为客户端无关。

## 11. net.core.rmem_max

## 12. net.core.wmem_max

## 13. net.ipv4.tcp_rmem

## 14. net.ipv4.tcp_wmem

以上4个参数决定了`socket buffer`大小，默认是几百KB，可以调大

## 附录

 

> 前言：
> Tcp/ip协议对网络编程的重要性，进行过网络开发的人员都知道，我们所编写的网络程序除了硬件，结构等限制，通过修改Tcp/ip内核参数也能得到很大的性能提升，
> 下面就列举一些Tcp/ip内核参数，解释它们的含义并通过修改来它们来优化我们的网络程序，主要是针对高并发情况。
> 这里网络程序主要指的是服务器端

------

### **1. fs.file-max**

> 最大可以打开的文件描述符数量，注意是整个系统。
> 在服务器中，我们知道每创建一个连接，系统就会打开一个文件描述符，所以，文件描述符打开的最大数量也决定了我们的最大连接数
> select在高并发情况下被取代的原因也是文件描述符打开的最大值，虽然它可以修改但一般不建议这么做，详情可见unp select部分。

------

### **2.net.ipv4.tcp_max_syn_backlog**

> Tcp syn队列的最大长度，在进行系统调用connect时会发生Tcp的三次握手，server内核会为Tcp维护两个队列，Syn队列和Accept队列，Syn队列是指存放完成第一次握手的连接，Accept队列是存放完成整个Tcp三次握手的连接，修改net.ipv4.tcp_max_syn_backlog使之增大可以接受更多的网络连接。
> 注意此参数过大可能遭遇到Syn flood攻击，即对方发送多个Syn报文端填充满Syn队列，使server无法继续接受其他连接
> 可参考此文http://tech.uc.cn/?p=1790

我们看下 man 手册上是如何说的：

> ```
>   The  behavior  of  the  backlog argument on TCP sockets changed with Linux 2.2.  Now it specifies the queue length for com‐
>    pletely established sockets waiting to be accepted, instead of the number of incomplete connection requests.   The  maximum
>    length  of  the  queue for incomplete sockets can be set using /proc/sys/net/ipv4/tcp_max_syn_backlog.  When syncookies are
>    enabled there is no logical maximum length and this setting is ignored.  See tcp(7) for more information.
> 
>   If the backlog argument is greater than the value in /proc/sys/net/core/somaxconn, then it is silently  truncated  to  that
>    value; the default value in this file is 128.  In kernels before 2.4.25, this limit was a hard coded value, SOMAXCONN, with
>    the value 128.
> ```

**自 Linux 内核 2.2 版本以后，backlog 为已完成连接队列的最大值，未完成连接队列大小以 /proc/sys/net/ipv4/tcp_max_syn_backlog 确定，但是已连接队列大小受 SOMAXCONN 限制，为 min(backlog, SOMAXCONN)**

------

### **3.net.ipv4.tcp_syncookies**

> 修改此参数可以有效的防范上面所说的syn flood攻击
> 原理：在Tcp服务器收到Tcp Syn包并返回Tcp Syn+ack包时，不专门分配一个数据区，而是根据这个Syn包计算出一个cookie值。在收到Tcp ack包时，Tcp服务器在根据那个cookie值检查这个Tcp ack包的合法性。如果合法，再分配专门的数据区进行处理未来的TCP连接。
> 默认为0，1表示开启

------

### **4.net.ipv4.tcp_keepalive_time**

> Tcp keepalive心跳包机制，用于检测连接是否已断开，我们可以修改默认时间来间断心跳包发送的频率。
> keepalive一般是服务器对客户端进行发送查看客户端是否在线，因为服务器为客户端分配一定的资源，但是Tcp 的keepalive机制很有争议，因为它们可耗费一定的带宽。
> Tcp keepalive详情见Tcp/ip详解卷1 第23章

------

### **5.net.ipv4.tcp_tw_reuse**

> 我的上一篇文章中写到了time_wait状态，大量处于time_wait状态是很浪费资源的，它们占用server的描述符等。
> 修改此参数，允许重用处于time_wait的socket。
> 默认为0，1表示开启

------

### **6.net.ipv4.tcp_tw_recycle**

> 也是针对time_wait状态的，该参数表示快速回收处于time_wait的socket。
> 默认为0，1表示开启

------

### **7.net.ipv4.tcp_fin_timeout**

> 修改time_wait状的存在时间，默认的2MSL
> 注意：time_wait存在且生存时间为2MSL是有原因的，见我上一篇博客为什么会有time_wait状态的存在，所以修改它有一定的风险，还是根据具体的情况来分析。

------

### **8.net.ipv4.tcp_max_tw_buckets**

> 所允许存在time_wait状态的最大数值，超过则立刻被清楚并且警告。

------

### **9.net.ipv4.ip_local_port_range**

> 表示对外连接的端口范围。

------

### **10.somaxconn**

> 前面说了Syn队列的最大长度限制，somaxconn参数决定Accept队列长度，在listen函数调用时backlog参数即决定Accept队列的长度，该参数太小也会限制最大并发连接数，因为同一时间完成3次握手的连接数量太小，server处理连接速度也就越慢。服务器端调用accept函数实际上就是从已连接Accept队列中取走完成三次握手的连接。
> Accept队列和Syn队列是listen函数完成创建维护的。
> /proc/sys/net/core/somaxconn修改









https://www.jianshu.com/p/3ecc99ebf566

https://www.cnblogs.com/alchemystar/p/13175276.html

nginx的backlog：https://www.04007.cn/article/323.html，https://www.jianshu.com/p/3ecc99ebf566

https://blog.csdn.net/jun2016425/article/details/81506353



Tcp的backlog：https://www.imooc.com/article/48429





https://www.awsok.com/possible-syn-flooding-on-port-80-sending-cookies%e9%97%ae%e9%a2%98%e5%a4%84%e7%90%86/

https://segmentfault.com/a/1190000008224853



https://www.cnblogs.com/study-everyday/p/9351831.html



Linux上TCP的几个内核参数调优：

https://www.cnblogs.com/alchemystar/p/13175276.html

https://www.cnblogs.com/study-everyday/p/9351831.html

https://blog.csdn.net/rain_qingtian/article/details/41864589

