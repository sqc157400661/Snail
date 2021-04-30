# 专题介绍

> 性能分析  性能优化 性能调优 排错 PProf Gops 



面向：需要有Go语言基础



[TOC]

### 题记：

在和女朋友或者男朋友在甜蜜的度假，突然收到系统报警，本来甜蜜的那啥，变成了惊心动魄的抢修。

### 经常发生的问题

2. 进程CPU使用率过高、内存占用不断增大（疑似泄露）
3. 临时内存大量申请后长时间不下降
4. goroutine泄露
4. goroutine数量暴涨
5. 服务器进程死掉或者不断重启
6. 进程还在，但是不提供服务
7. 并发多的时候，程序才发生异常
8. 重构系统后，上线后发现性能不佳
9. 某次迭代发布后的数小时内出现了应用程序无法提供服务的情况

如何避免这些问题那？平时我们应该做些什么？

面对这些问题，除在平时要做好各类防护外，在出现问题时，应如何排查呢？

### 专题的目标

1. 掌握如何进行性能指标分析
2. 学会使用常见性能分析工具
3. 拥有快速定位问题的思路和手段

### 专题包含的内容

1. 性能剖析工具PProf
   	1. 压力测试工具ab
    2. PProf简介
    3. PProf的简单使用
    4. 通过交互式终端使用
    5. 可视化界面
    6. 与性能测试结合做剖析
    7. 小结
    8. 实例1-排查CPU占用过高问题
    9. 实例2-排查内存占用过高问题
    10. 实例3-排查频繁GC问题
    11. 实例4-排查协程泄漏问题
    12. 实例5-排查锁竞争问题
    13. 实例6-排查阻塞问题
2. fgprof   https://xargin.com/go-perf-optimization/  https://github.com/felixge/fgprof
3. data race
4. GODEBUG工具
5. 进程诊断工具Gops
6. 逃逸分析
7. 常见线上问题举例说明





参考：

https://github.com/google/pprof/blob/master/doc/README.md

https://www.php.cn/manual/view/35260.html



trace：https://blog.csdn.net/qq_30549833/article/details/89381790

golang pprof 实战https://blog.wolfogre.com/posts/go-ppof-practice/



https://xargin.com/go-perf-optimization/





https://www.kancloud.cn/aceld/golang/1958304



https://colobu.com/2019/08/20/use-pprof-to-compare-go-memory-usage/





https://studygolang.com/articles/30507