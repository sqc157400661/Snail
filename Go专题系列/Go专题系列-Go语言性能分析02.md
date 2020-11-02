# 性能剖析工具PProf

## 第一节：压力测试工具ab

### ab工具介绍

ab是apache自带的压力测试工具。ab非常实用，它不仅可以对apache服务器进行网站访问压力测试，也可以对或其它类型的服务器进行压力测试。比如nginx、tomcat、IIS等

### 安装

```csharp
yum -y install httpd-tools # centos，redhat

apt-get install apache2-utils   #ubuntu Debian 
```

### 使用说明

```swift
# 格式
ab -n1000 -c 10 http:/www.snail.com/ 
```

常用参数说明：

```
-n  即requests，用于指定压力测试总共的执行次数。
-c  即concurrency，用于指定压力测试的并发数。
-t  即timelimit，等待响应的最大时间(单位：秒)。
```

```
# -n发出800个请求，-c模拟800并发，相当800人同时访问，后面是测试url
ab -n 800 -c 800  http://192.168.0.10/ 

#在60秒内发请求，一次100个并发请求。
ab -t 60 -c 100 http://192.168.0.10/ 
```

更多使用方法详见 [ab 官方文档](http://httpd.apache.org/docs/2.0/programs/ab.html)

### 结果说明

```
Server Software:        Apache          #服务器软件
Server Hostname:        www.taoquan.ink #域名
Server Port:            80              #请求端口号

Document Path:          /               #文件路径
Document Length:        40888 bytes     #页面字节数

Concurrency Level:      10              #请求的并发数
Time taken for tests:   27.300 seconds  #总访问时间
Complete requests:      1000            #请求成功数量
Failed requests:        0               #请求失败数量
Write errors:           0
Total transferred:      41054242 bytes  #请求总数据大小（包括header头信息）
HTML transferred:       40888000 bytes  #html页面实际总字节数
Requests per second:    36.63 [#/sec] (mean)  #每秒多少请求，这个是非常重要的参数数值，服务器的吞吐量
Time per request:       272.998 [ms] (mean)     #用户平均请求等待时间 
Time per request:       27.300 [ms] (mean, across all concurrent requests) # 服务器平均处理时间，也就是服务器吞吐量的倒数                  
Transfer rate:          1468.58 [Kbytes/sec] received  #每秒获取的数据长度

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:       43   47   2.4     47      53
Processing:   189  224  40.7    215     895
Waiting:      102  128  38.6    118     794
Total:        233  270  41.3    263     945

Percentage of the requests served within a certain time (ms)
  50%    263    #50%用户请求在263ms内返回
  66%    271    #66%用户请求在271ms内返回
  75%    279    #75%用户请求在279ms内返回
  80%    285    #80%用户请求在285ms内返回
  90%    303    #90%用户请求在303ms内返回
  95%    320    #95%用户请求在320ms内返回
  98%    341    #98%用户请求在341ms内返回
  99%    373    #99%用户请求在373ms内返回
 100%    945 (longest request)
```

