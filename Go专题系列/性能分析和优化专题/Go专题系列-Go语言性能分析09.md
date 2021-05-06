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

# 注意，列表中有个程序名称后面带了个 *，表示该程序加入了 gops 的诊断分析代码。
```

- 在上述输出中，第3行的输出结果中包含了一个`*`符号，这代表着该Go进程包含了agent，因此它可以启用更强大的诊断功能，包括当前堆栈跟踪、Go版本、内存统计信息，等等。
- 不包含*符号，这意味着它是一个普通的Go程序，即没有植入agent，只能使用最基本的功能，也就是说无法执行`gops memstats`、`gops pprof-heap`等所有类似于 `gops <cmd> <pid|addr> ...` 的子命令。

ps:`agent.Options` 参数说明：

```
// Code reference: github.com/google/gops/agent/agent.go:42

// Options allows configuring the started agent.
type Options struct {
	// Addr is the host:port the agent will be listening at.
	// Optional.
	Addr string

	// ConfigDir is the directory to store the configuration file,
	// PID of the gops process, filename, port as well as content.
	// Optional.
	ConfigDir string

	// ShutdownCleanup automatically cleans up resources if the
	// running process receives an interrupt. Otherwise, users
	// can call Close before shutting down.
	// Optional.
	ShutdownCleanup bool
}
```

- Addr

  可选。为远程分析服务提供监听地址，例如: `:9119`。配置了该项，那我们可以在本机查看分析远程服务器上的 Go 程序，非常有帮助。

- ConfigDir

  可选。用于存放统计数据（go进程信息）和配置的目录，默认为当前用户的主目录。也可以通过环境变量`GOPS_CONFIG_DIR`设置。具体参考代码：

  ```
  const gopsConfigDirEnvKey = "GOPS_CONFIG_DIR"
  
  func ConfigDir() (string, error) {
  	if configDir := os.Getenv(gopsConfigDirEnvKey); configDir != "" {
  		return configDir, nil
  	}
  
  	if runtime.GOOS == "windows" {
  		return filepath.Join(os.Getenv("APPDATA"), "gops"), nil
  	}
  	homeDir := guessUnixHomeDir()
  	if homeDir == "" {
  		return "", errors.New("unable to get current user home directory: os/user lookup failed; $HOME is empty")
  	}
  	return filepath.Join(homeDir, ".config", "gops"), nil
  }
  
  func guessUnixHomeDir() string {
  	usr, err := user.Current()
  	if err == nil {
  		return usr.HomeDir
  	}
  	return os.Getenv("HOME")
  }
  ```

- ShutdownCleanup

  可选。设置为 `true`，则在程序关闭时会自动清理数据（ConfigDir中的文件）。

## 常用命令

gops工具包含了大量的分析命令，我们可以通过gops help进行查看：

```
$ gops help
gops is a tool to list and diagnose Go processes.

Usage:
gops <cmd> <pid|addr> ...
gops <pid> # displays process info
gops help  # displays this help message

Commands:
stack      Prints the stack trace.
gc         Runs the garbage collector and blocks until successful.
setgc      Sets the garbage collection target percentage.
memstats   Prints the allocation and garbage collection stats.
version    Prints the Go version used to build the program.
stats      Prints runtime stats.
trace      Runs the runtime tracer for 5 secs and launches "go tool trace".
pprof-heap Reads the heap profile and launches "go tool pprof".
pprof-cpu  Reads the CPU profile and launches "go tool pprof".
```

接下来将针对几个常用的分析功能进行简要分析。

### 1、查看指定进程信息

用法: `gops <pid>` 查看本机指定 `PID` Go 程序的基本信息

```
$ gops 1556
parent PID: 3725
threads:    7
memory usage:   0.042%
cpu usage:  0.003%
username:   eddycjy
cmd+args:   /var/folders/jm/pk20jr_s74x49kqmyt87n2800000gn/T/go-build943691423/b001/exe/main
elapsed time:   10:56
local/remote:   127.0.0.1:59369 <-> :0 (LISTEN)
local/remote:   *:6060 <-> :0 (LISTEN)
```

- 获取Go进程的概要信息，包括父级PID、线程数、内存或CPU使用率、运行者的账户名、进程的启动命令行参数、启动后所经过的时间，以及gops的agent监听信息（若没有植入agent，则没有这项信息）。
- `local/remote:   *:6060 <-> :0 (LISTEN)` 是 `gops/agent` 提供的服务



### 2、查看调用栈信息

用法: `gops stack (<pid>|<addr>)` 用于显示程序所有堆栈信息，包括每个 goroutine 的堆栈信息、运行状态、运行时长等。也可用于分析调用链路。

```
$ gops stack 516
goroutine 7 [running]:
runtime/pprof.writeGoroutineStacks(0x77d4c0, 0xc0000c0000, 0x30, 0x95c380)
        C:/ThsSoftware/Golang_ths/ths/src/runtime/pprof/pprof.go:679 +0xa4
runtime/pprof.writeGoroutine(0x77d4c0, 0xc0000c0000, 0x2, 0x1c0cf8a8, 0x9207f9987614b1cb)
        C:/ThsSoftware/Golang_ths/ths/src/runtime/pprof/pprof.go:668 +0x4b
runtime/pprof.(*Profile).WriteTo(0x950aa0, 0x77d4c0, 0xc0000c0000, 0x2, 0xc0000c0000, 0x0)
        C:/ThsSoftware/Golang_ths/ths/src/runtime/pprof/pprof.go:329 +0x3e1
github.com/google/gops/agent.handle(0x2470008, 0xc0000c0000, 0xc0000142c9, 0x1, 0x1, 0x0, 0x0)
        D:/www/Snail/Go娑撴捇顣界化璇插灙/code/Go鐠囶叀鈻堥幀褑鍏橀崚鍡樼€▒/gops/vendor/github.com/google/gops/agent/agent.go:189 +0x1b2
............
```

### 3、查看内存使用情况

用法: `gops memstats (<pid>|<addr>)` 查看程序的内存统计信息

获取Go运行时的当前内存使用情况，主要是runtime.MemStats的相关字段信息。

```
$  gops memstats 516
alloc: 1.16MB (1221176 bytes)
total-alloc: 1.16MB (1221176 bytes)
sys: 6.38MB (6687096 bytes)
lookups: 0
mallocs: 916
frees: 21
heap-alloc: 1.16MB (1221176 bytes)
heap-sys: 3.88MB (4063232 bytes)
heap-idle: 2.10MB (2203648 bytes)
heap-in-use: 1.77MB (1859584 bytes)
heap-released: 2.07MB (2170880 bytes)
heap-objects: 895
stack-in-use: 128.00KB (131072 bytes)
stack-sys: 128.00KB (131072 bytes)
stack-mspan-inuse: 11.55KB (11832 bytes)
stack-mspan-sys: 16.00KB (16384 bytes)
stack-mcache-inuse: 3.33KB (3408 bytes)
stack-mcache-sys: 16.00KB (16384 bytes)
other-sys: 787.17KB (806066 bytes)
gc-sys: 206.12KB (211072 bytes)
next-gc: when heap-alloc >= 4.27MB (4473924 bytes)
last-gc: -
gc-pause-total: 0s
gc-pause: 0
gc-pause-end: 0
num-gc: 0
num-forced-gc: 0
gc-cpu-fraction: 0
enable-gc: true
debug-gc: false
```

### 4 、查看运行时信息

用法: `gops stats (<pid>|<addr>)` 获取Go运行时的基本信息，包括当前的goroutine数量、系统线程、GOMAXPROCS数值及当前系统的CPU核数。

```
$  gops stats 516
goroutines: 3
OS threads: 5
GOMAXPROCS: 2
num CPU: 2
```

### 5、查看trace信息

用法: `gops trace (<pid>|<addr>)` 追踪程序运行5秒，生成可视化报告   与go tool trace的作用基本一致。

```
$  gops trace 516
Tracing now, will take 5 secs...
Trace dump saved to: C:\Users\VIRUSE~1.V-D\AppData\Local\Temp\trace319644651
2020/11/18 13:42:45 Parsing trace...
2020/11/18 13:42:45 Splitting trace...
2020/11/18 13:42:46 Opening browser. Trace viewer is listening on http://127.0.0.1:55857
```

### 6、查看profile信息

用法: `gops pprof-cpu (<pid>|<addr>)` 调用并展示 `go tool pprof` 工具中关于 CPU 的性能分析数据，操作与 `pprof` 一致

用法: `gops pprof-heap (<pid>|<addr>)` 调用并展示 `go tool pprof` 工具中关于 heap 的性能分析数据，操作与 `pprof` 一致。

### 7、其他

1. `gops gc (<pid>|<addr>)` 查看指定程序的垃圾回收(GC)信息
2. `gops setgc (<pid>|<addr>)` 设定指定程序的 GC 目标百分比
3. `gops version (<pid>|<addr>)` 查看指定程序构建时的 Go 版本号
4. `gops tree` 以目录树的形式展示所有 Go 程序。



## 小结：

gops工具，是工具的集大成者





## 参考

1. https://github.com/google/gops
2. [google/gops源码分析](https://github.com/XanthusL/blog-gen/blob/master/content/post/gops.md)