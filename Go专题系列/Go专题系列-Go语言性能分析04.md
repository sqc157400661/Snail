# 实例分析说明

## 第二例概述：XXXXX

### 场景再现：

一个并发活动，并发量大于1万，突然出现了问题，页面半天打不开，打开了半天下不了单，cpu涨了又跌跌了又涨，领导发飙，运营骂娘。于是技术团队开始定位bug，看了好几圈都没有发现问题，重启也没有解决问题....



### 分析问题，提出假设和疑问

为什么单独CPU消耗这么大？为什么还有波动？

#### 猜测和假设出CPU暴涨的原因：

1. 某段代码逻辑计算密度过大
2. GC压力过大？（小对象太多？）
3. 依赖接口请求延迟，导致在高并发环境下请求积压越来越多？

有了初步的分析和假设，下一步我们来一起定位问题

### 定位问题，使用趁手的工具

#### 1、用`pprof`工具定位







参考：

1. https://mp.weixin.qq.com/s/wXzGZlvF6fCXjcGFzG74Xw

2. https://mp.weixin.qq.com/s?__biz=MzAxMTA4Njc0OQ==&mid=2651439006&idx=1&sn=0db8849336cc4172c663a574212ea8db&chksm=80bb616cb7cce87a1dc529e6c8bdcf770e293fc4ce67ede8e1908199480534c39f79803038e3&scene=21#wechat_redirect

3. https://mp.weixin.qq.com/s?__biz=MzAxMTA4Njc0OQ==&mid=2651438884&idx=1&sn=809c8d3041bf6913fa8bc1e57e6cb5c6&chksm=80bb61d6b7cce8c04ed610fe2fd333ff02afe24baeadf3d98ae84f7a8cef247687518239bc5f&scene=21#wechat_redirect

4. http://mp.weixin.qq.com/mp/homepage?__biz=MzAxMTA4Njc0OQ==&hid=11&sn=ebf226cfb8ac6f5140fe133c0bd80dd1&scene=18#wechat_redirect

   