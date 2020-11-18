# Go进程诊断工具gops

在类UNIX系统中，我们可以使用ps命令查看系统当前正在运行的进程信息，快速查看某些进程的运行情况和状态。

想必 Java 的开发者没有不知道或者没用过 **jps** 这个命令的，这个命令是用来在主机上查看和分析 Java 进程的。

## gops简介

那么 Go 语言有没有像 jps 这样的工具呢？当然有，不仅有，而且还是 Google 自己出品的，官方认证（这种问题 Google 不可能自己想不到啊）。名称也跟 jps 很像，叫 **gops**。通过它可以查看并诊断当前系统中Go程序的运行情况及状态，属于常用工具。

常用功能：

可以查看：

- 当前有哪些go语言进程，哪些使用gops的go进程
- 进程的概要信息
- 进程的调用栈
- 进程的内存使用情况
- 构建程序的Go版本
- 运行时统计信息

可以获取：

- trace
- cpu profile和memory profile

还可以：

- 让进程进行1次GC
- 设置GC百分比

## 基本使用

**gops** 并不包含在官方安装包中，不属于标准工具。需要手动获取。

```
go get -u github.com/google/gops
```

windows安装：

```
go install github.com/google/gops
// 下载完成后安装对应包 会生成 gops.exe 文件
// 请放到系统环境变量里面 如果运行install正常来说应该生成在%GOPATH%/bin/下面
// gops -help检查一下是否安装成功
```

写入如下启动代码：

```
package main

import (
	"github.com/google/gops/agent"
	"log"
	"net/http"
)

func main() {
	// 创建并监听 gops agent，gops 命令会通过连接 agent 来读取进程信息
	// 若需要远程访问，可配置 agent.Options{Addr: "0.0.0.0:6060"}，否则默认仅允许本地访问
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatalf("agent.Listen err: %v", err)
	}

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`Hello Gops`))
	})
	http.ListenAndServe(":6060", http.DefaultServeMux)
}
```

启动该程序，并在命令行执行gops命令进行查看：

```
$ gops
4252 5656 go_build_snailpprof.exe  go1.13 D:\www\Snail\exe\go_build_snailpprof.exe
5204 4924 gops.exe                 go1.13 C:\Users\viruser.v-desktop\go\bin\gops.exe
1556 4320 go_build_gopsdemo.exe  * go1.13 D:\www\Snail\exe\go_build_gopsdemo.exe
```

结果说明：

```
PID，PPID，程序名称，编译使用的 Go 版本号，程序路径
```

- 在上述输出中，第3行的输出结果中包含了一个`*`符号，这代表着该Go进程包含了agent，因此它可以启用更强大的诊断功能，包括当前堆栈跟踪、Go版本、内存统计信息，等等。
- 不包含*符号，这意味着它是一个普通的Go程序，即没有植入agent，只能使用最基本的功能。







## 参考

1. https://github.com/google/gops