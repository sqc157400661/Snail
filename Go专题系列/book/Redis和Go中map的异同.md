## Redis和Go中map的异同

Redis和Go中的map实现，有很多相似之处。这里做一个总结，方便大家深入理解和记忆。先来两张图：

Redis map数据类型：（来自[《Redis设计与实现》](http://redisbook.com/index.html)）![Redis map类型结构](D:\www\Snail\Go专题系列\book\images\19f4f05e7b4dc9a0e1a2d631be8fb8b6.png)

Go map类型结构:（来自[饶大博客](https://qcrao.com/2019/05/22/dive-into-go-map/)）
![Go map类型结构](D:\www\Snail\Go专题系列\book\images\b0ee1586baa21740d2568debdfd27f9f.png)

## 数据结构

- 相同：内部两个哈希表，用于扩容，但Go中叫做buckets和oldbuckets，Redis中是一个数组，大小为2
- 不同：层次不同。 参见上面的图，Redis第二层存储了子表的信息，第三层作为子表，存储的是实际数据的地址；Go实际只有三层。这导致两种实现后续功能的差异，如：是否支持缩容。
- 不同：size、used存储方式。Go在顶层结构中存储了B字段，表示有2^B个bucket，并存储了count字段表示已有数据个数；Redis在第二层(每个字表信息)中
- 不同：k-v排列方式。Go：8*key+8*val作为一个bucket，后面可以链式挂接更多overflow的bucket，Redis：key+val作为一个dictEntity，后面可以链式挂接更多dictEntity

## 哈希方式

- 相同：根据不同类型，调用不同的hash方法后，求余得到索引
- 不同：Redis：得到bucket索引后即得到bucket数据，二Go还需要再根据tophash->key的查找

## 冲突解决方式

- 相同：拉链法
- 不同：Redis的链表直接存在每个数据(dictEntity)后，Go由8个k-v组成一个bucket，然后再挂接overflow bucket

## rehash

- 相同：装载因子的概念
- 相同：渐进式扩容 下文详细描述
- 不同：触发扩容的时机 Go在插入新key时检测装载因子和拉链长度，Redis在增删查改时都会检查是否需要rehash
- 不同：触发扩容的条件 Go：bucket内(可以存储8个k-v)的平均个数超过6.5或单个bucket的overflow超过bucket数会进行扩容；Redis：每个key平均存储了一个数据，则进行扩容；每个key平均存储了不到0.1个数据，则进行缩容

> 思考：
>
> > 假如Go的数据主要集中在一个bucket里，其overflow很长，其他bucket的数据很少，这时应该触发rehash吗？
> > 这要分两种情况：
> >
> > - 如果该bucket包括overflow中的数据量比较满，那么map整体的数据量也接近6.6/bucket了，rehash扩容后，hash后的key被重新打散，数据会被重新分配到其他bucket中；
> > - 如果该overflow中的位置比较空，rehash后bucket数量不变，该bucket中的数据被重新依次填入到新的bucket中，空位消失，overflow数量也就减少了

## 渐进式扩容

- 相同：第一次先分配空间，后面再渐进搬迁
- 不同：Go只在增删操作时搬迁，Redis在增删查改操作时都会进行渐进搬迁操作

## 缩容

- 不同：Go只保存了当前的bucket size，新bucket一定是旧bucket大小的两倍，不支持缩容。Redis的缩容：两个子表记录了自己的大小，缩容即扩容的逆过程

## 代码组织方式

- 不同：Go编译器会做很多trick的处理，比如给bmap填充真实的字段和数据，而Redis使用C编写，代码一目了然