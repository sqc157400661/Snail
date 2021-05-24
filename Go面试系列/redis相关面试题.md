# Redis基础

## Redis中的数据结构

### 1. 底层数据结构, 与`Redis Value Type`之间的关系

对于Redis的使用者来说, Redis作为Key-Value型的内存数据库, 其Value有多种类型.

- String
- Hash
- List
- Set
- ZSet

这些Value的类型, 只是"Redis的用户认为的, Value存储数据的方式". 而在具体实现上,各个Type的Value到底如何存储, 这对于Redis的使用者来说是不公开的.
举个粟子: 使用下面的命令创建一个Key-Value

```
SET "Hello" "World"
```

对于Redis的使用者来说, Hello 这个Key, 对应的Value是String类型, 其值为五个ASCII字符组成的二进制数据. 但具体在底层实现上, 这五个字节是如何存储的, 是不对用户公开的. 即, Value的Type, 只是表象, 具体数据在内存中以何种数据结构存放, 这对于用户来说是不必要了解的.

Redis对使用者暴露了五种 `Value Type`, 其底层实现的数据结构有8种, 分别是:

- `SDS - simple synamic string` - 支持自动动态扩容的字节数组
- `list` - 平平无奇的链表
- `dict` - 使用双哈希表实现的, 支持平滑扩容的字典
- `zskiplist` - 附加了后向指针的跳跃表
- `intset` - 用于存储整数数值集合的自有结构
- `ziplist` - 一种实现上类似于TLV, 但比TLV复杂的, 用于存储任意数据的有序序列的
  数据结构
- `quicklist` - 一种以ziplist作为结点的双链表结构, 实现的非常苟
- `zipmap` - 一种用于在小规模场合使用的轻量级字典结构

而衔接"底层数据结构"与"Value Type"的桥梁的, 则是Redis实现的另外一种数据结
构: `redisObject`. Redis中的Key与Value在表层都是一个 `redisObject` 实例, 故该结构有
所谓的"类型", 即是 ValueType. 对于每一种Value Type 类型的`redisObject` , 其底层至
少支持两种不同的底层数据结构来实现. 以应对在不同的应用场景中, Redis的运行效率,
或内存占用.



### 2. 底层数据结构

#### 2.1 SDS - simple dynamic string

这是一种用于存储二进制数据的一种结构, 具有动态扩容的特点. 其实现位于src/sds.h
与src/sds.c中, 其关键定义如下:

```
typedef char *sds;

/* Note: sdshdr5 is never used, we just access the flags byte directly.
 * However is here to document the layout of type 5 SDS strings. */
struct __attribute__ ((__packed__)) sdshdr5 {
    unsigned char flags; /* 3 lsb of type, and 5 msb of string length */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr8 {
    uint8_t len; /* used */
    uint8_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr16 {
    uint16_t len; /* used */
    uint16_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr32 {
    uint32_t len; /* used */
    uint32_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr64 {
    uint64_t len; /* used */
    uint64_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
```

#### SDS的总体概览如下图:

![sds](D:\www\Snail\Go面试系列\images\668722-20180910183925685-224282854.png)

其中`sdshdr`是头部, `buf`是真实存储用户数据的地方. 另外注意, 从命名上能看出来, 这个数据结构除了能存储二进制数据, 显然是用于设计作为字符串使用的, 所以在buf中, 用户数据后总跟着一个\0. 即图中 "数据" + "\0" 是为所谓的buf

SDS有五种不同的头部. 其中sdshdr5实际并未使用到. 所以实际上有四种不同的头部,分别如下:

![sdshdr](D:\www\Snail\Go面试系列\images\668722-20180910183940043-66166526.png)

1. `len`分别以`uint8`, `uint16`, `uint32`, `uint64`表示用户数据的长度(不包括末尾的\0)
2. `alloc`分别以`uint8`, `uint16`, `uint32`, `uint64`表示整个SDS, 除过头部与末尾的\0, 剩余的字节数.
3. flag始终为一字节, 以低三位标示着头部的类型, 高5位未使用.

当在程序中持有一个SDS实例时, 直接持有的是数据区的头指针, 这样做的用意是: 通过
这个指针, 向前偏一个字节, 就能取到flag, 通过判断flag低三位的值, 能迅速判断: 头部的
类型, 已用字节数, 总字节数, 剩余字节数. 这也是为什么sds类型即是char *指针类型别
名的原因.
创建一个SDS实例有三个接口, 分别是:



```
 所有创建sds实例的接口, 都不会额外分配预留内存空间
```

```
1 // 创建一个不含数据的sds:
```

(^2) // 头部 3 字节 sdshdr8
(^3) // 数据区 0 字节
(^4) // 末尾 \0 占一字节
(^5) sds sdsempty(void);
6 // 带数据创建一个sds:
7 // 头部 按initlen的值, 选择最小的头部类型
8 // 数据区 从入参指针init处开始, 拷贝initlen个字节
(^9) // 末尾 \0 占一字节
(^10) sds sdsnewlen(const void *init, size_t initlen);
(^11) // 带数据创建一个sds:
(^12) // 头部 按strlen(init)的值, 选择最小的头部类型
(^13) // 数据区 入参指向的字符串中的所有字符, 不包括末尾 \0
14 // 末尾 \0 占一字节
15 sds sdsnew(const char *init);


```
 sdsnewlen用于带二进制数据创建sds实例, sdsnew用于带字符串创建sds实例. 接
口返回的sds可以直接传入libc中的字符串输出函数中进行操作, 由于无论其中存储的
是用户的二进制数据, 还是字符串, 其末尾都带一个\0, 所以至少调用libc中的字符串
输出函数是安全的.
在对SDS中的数据进行修改时, 若剩余空间不足, 会调用sdsMakeRoomFor函数用于扩
容空间, 这是一个很低级的API, 通常情况下不应当由SDS的使用者直接调用. 其实现中核
心的几行如下:
```

#### 可以看到, 在扩充空间时

```
 先保证至少有addlen可用
 然后再进一步扩充, 在总体占用空间不超过阈值 SDS_MAC_PREALLOC 时, 申请空间再
翻一倍. 若总体空间已经超过了阈值, 则步进增⻓ SDS_MAC_PREALLOC. 这个阈值的
默认值为  1024 * 1024
SDS也提供了接口用于移除所有未使用的内存空间. sdsRemoveFreeSpace , 该接口没有
间接的被任何SDS其它接口调用, 即默认情况下, SDS不会自动回收预留空间. 在SDS的
使用者需要节省内存时, 由使用者自行调用:
```

#### 总结:

```
 SDS除了是某些Value Type的底层实现, 也被大量使用在Redis内部, 用于替代C-
Style字符串. 所以默认的创建SDS实例接口, 不分配额外的预留空间. 因为多数字符
串在程序运行期间是不变的. 而对于变更数据区的API, 其内部则是调用了
sdsMakeRoomFor, 每一次扩充空间, 都会预留大量的空间. 这样做的考量是: 如果一
个SDS实例中的数据被变更了, 那么很有可能会在后续发生多次变更.
 SDS的API内部不负责清除未使用的闲置内存空间, 因为内部API无法判断这样做的合
适时机. 即便是在操作数据区的时候导致数据区占用内存减少时, 内部API也不会清除
闲置内在空间. 清除闲置内存空间责任应当由SDS的使用者自行担当.
 用SDS替代C-Style字符串时, 由于其头部额外存储了数据区的⻓度信息, 所以字符
串的求⻓操作时间复杂度为O(1)
```

```
2.2 list
```

```
1 sds sdsMakeRoomFor(sds s, size_t addlen) {
```

(^2) ...
(^3) /* Return ASAP if there is enough space left. */
(^4) if (avail >= addlen) return s;
56 len = sdslen(s);
7 sh = (char*)s-sdsHdrSize(oldtype);
8 newlen = (len+addlen);
(^9) if (newlen < SDS_MAX_PREALLOC)
(^10) newlen *= 2;
(^11) else
(^12) newlen += SDS_MAX_PREALLOC;
(^13) ...
(^14) }
1 sds sdsRemoveFreeSpace(sds s);


```
这是普通的链表实现, 链表结点不直接持有数据, 而是通过void *指针来间接的指向数据.
其实现位于 src/adlist.h与src/adlist.c中, 关键定义如下:
```

#### 其内存布局如下图所示:

```
这是一个平平无奇的链表的实现. list在Redis除了作为一些Value Type的底层实现外, 还
广泛用于Redis的其它功能实现中, 作为一种数据结构工具使用. 在list的实现中, 除了基
本的链表定义外, 还额外增加了:
 迭代器 listIter的定义, 与相关接口的实现.
 由于list中的链表结点本身并不直接持有数据, 而是通过value字段, 以void *指针的形
式间接持有, 所以数据的生命周期并不完全与链表及其结点一致. 这给了list的使用者
相当大的灵活性. 比如可以多个结点持有同一份数据的地址. 但与此同时, 在对链表进
行销毁, 结点复制以及查找匹配时, 就需要list的使用者将相关的函数指针赋值于
list.dup, list.free, list.match字段.
```

```
2.3 dict
dict是Redis底层数据结构中实现最为复杂的一个数据结构, 其功能类似于C++标准库中
的std::unordered_map, 其实现位于 src/dict.h 与 src/dict.c中, 其关键定义如下:
```

```
1 typedef struct listNode {
```

(^2) struct listNode *prev;
3 struct listNode *next;
4 void *value;
(^5) } listNode;
(^67) typedef struct listIter {
(^8) listNode *next;
(^9) int direction;
(^10) } listIter;
(^1112) typedef struct list {
13 listNode *head;
14 listNode *tail;
15 void *(*dup)(void *ptr);
(^16) void (*free)(void *ptr);
(^17) int (*match)(void *ptr, void *key);
(^18) unsigned long len;
(^19) } list;
1 typedef struct dictEntry {
2 void *key;
(^3) union {
(^4) void *val;


#### 其内存布局如下所示:

```
5 uint64_t u64;
6 int64_t s64;
```

(^7) double d;
(^8) } v;
(^9) struct dictEntry *next;
(^10) } dictEntry;
(^1112) typedef struct dictType {
13 uint64_t (*hashFunction)(const void *key);
14 void *(*keyDup)(void *privdata, const void *key);
15 void *(*valDup)(void *privdata, const void *obj);
(^16) int (*keyCompare)(void *privdata, const void *key1, const void *key2);
(^17) void (*keyDestructor)(void *privdata, void *key);
(^18) void (*valDestructor)(void *privdata, void *obj);
(^19) } dictType;
(^20) /* This is our hash table structure. Every dictionary has two of this as
we
21
22 * implement incremental rehashing, for the old to the new table. */
23 typedef struct dictht {
24 dictEntry **table;
(^25) unsigned long size;
(^26) unsigned long sizemask;
(^27) unsigned long used;
(^28) } dictht;
(^2930) typedef struct dict {
31 dictType *type;
32 void *privdata;
33 dictht ht[2];
(^34) long rehashidx; /* rehashing not in progress if rehashidx == -1 */
(^35) unsigned long iterators; /* number of iterators currently running */
(^36) } dict;
(^3738) /* If safe is set to 1 this is a safe iterator, that means, you can call

* dictAdd, dictFind, and other functions against the dictionary even
  while
  39
  40 * iterating. Otherwise it is a non safe iterator, and only dictNext()
  41 * should be called while iterating. */
  42 typedef struct dictIterator {
  (^43) dict *d;
  (^44) long index;
  (^45) int table, safe;
  (^46) dictEntry *entry, *nextEntry;
  (^47) /* unsafe iterator fingerprint for misuse detection. */
  48 long long fingerprint;
  49 } dictIterator;


 dict中存储的键值对, 是通过dictEntry这个结构间接持有的, k通过指针间接持有键, v
通过指针间接持有值. 注意, 若值是整数值的话, 是直接存储在v字段中的, 而不是间接
持有. 同时next指针用于指向, 在bucket索引值冲突时, 以链式方式解决冲突, 指向同
索引的下一个dictEntry结构.
 传统的哈希表实现, 是一块连续空间的顺序表, 表中元素即是结点. 在dictht.table中,
结点本身是散布在内存中的, 顺序表中存储的是dictEntry的指针
 哈希表即是dictht结构, 其通过table字段间接的持有顺序表形式的bucket, bucket的
容量存储在size字段中, 为了加速将散列值转化为bucket中的数组索引, 引入了
sizemask字段, 计算指定键在哈希表中的索引时, 执行的操作类似于dict->type-

>hashFunction(键) & dict->ht[x].sizemask. 从这里也可以看出来, bucket的容量
>适宜于为2的幂次, 这样计算出的索引值能覆盖到所有bucket索引位.
> dict即为字典. 其中type字段中存储的是本字典使用到的各种函数指针, 包括散列函
>数, 键与值的复制函数, 释放函数, 以及键的比较函数. privdata是用于存储用户自定
>义数据. 这样, 字典的使用者可以最大化的自定义字典的实现, 通过自定义各种函数实
>现, 以及可以附带私有数据, 保证了字典有很大的调优空间.
> 字典为了支持平滑扩容, 定义了ht[2]这个数组字段. 其用意是这样的:
> 一般情况下, 字典dict仅持有一个哈希表dictht的实例, 即整个字典由一个bucket实现.
> 随着插入操作, bucket中出现冲突的概率会越来越大, 当字典中存储的结点数目, 与
>bucket数组⻓度的比值达到一个阈值(1:1)时, 字典为了缓解性能下降, 就需要扩容
> 扩容的操作是平滑的, 即在扩容时, 字典会持有两个dictht的实例, ht[0]指向旧哈希表,
>ht[1]指向扩容后的新哈希表. 平滑扩容的重点在于两个策略:
> 后续每一次的插入, 替换, 查找操作, 都插入到ht[1]指向的哈希表中
> 每一次插入, 替换, 查找操作执行时, 会将旧表ht[0]中的一个bucket索引位持有的结点
>链表, 迁移到ht[1]中去. 迁移的进度保存在rehashidx这个字段中.在旧表中由于冲突而被
>链接在同一索引位上的结点, 迁移到新表后, 可能会散布在多个新表索引中去.
> 当迁移完成后, ht[0]指向的旧表会被释放, 之后会将新表的持有权转交给ht[0], 再重置
>ht[1]指向NULL
> 这种平滑扩容的优点有两个:
> 平滑扩容过程中, 所有结点的实际数据, 即dict->ht[0]->table[rehashindex]->k与
>dict->ht[0]->table[rehashindex]->v分别指向的实际数据, 内存地址都不会变化. 没有


```
发生键数据与值数据的拷⻉或移动, 扩容整个过程仅是各种指针的操作. 速度非常快
 扩容操作是步进式的, 这保证任何一次插入操作都是顺畅的, dict的使用者是无感知的.
若扩容是一次性的, 当新旧bucket容量特别大时, 迁移所有结点必然会导致耗时陡增.
```

```
除了字典本身的实现外, 其中还顺带实现了一个迭代器, 这个迭代器中有字段safe以标示
该迭代器是"安全迭代器"还是"非安全迭代器", 所谓的安全与否, 指是的这种场景: 设想在
运行迭代器的过程中, 字典正处于平滑扩容的过程中. 在平滑扩容的过程中时, 旧表一个
索引位上的, 由冲突而链起来的多个结点, 迁移到新表后, 可能会散布到新表的多个索引
位上. 且新的索引位的值可能比旧的索引位要低.
遍历操作的重点是, 保证在迭代器遍历操作开始时, 字典中持有的所有结点, 都会被遍历
到. 而若在遍历过程中, 一个未遍历的结点, 从旧表迁移到新表后, 索引值减小了, 那么就
可能会导致这个结点在遍历过程中被遗漏.
所以, 所谓的"安全"迭代器, 其在内部实现时: 在迭代过程中, 若字典正处于平滑扩容过程,
则暂停结点迁移, 直至迭代器运行结束. 这样虽然不能保证在迭代过程中插入的结点会被
遍历到, 但至少保证在迭代起始时, 字典中持有的所有结点都会被遍历到.
这也是为什么dict结构中有一个iterators字段的原因: 该字段记录了运行于该字典上的
安全迭代器的数目. 若该数目不为0, 字典是不会继续进行结点迁移平滑扩容的.
下面是字典的扩容操作中的核心代码, 我们以插入操作引起的扩容为例:
先是插入操作的外部逻辑:
1.如果插入时, 字典正处于平滑扩容过程中, 那么无论本次插入是否成功, 先迁移一个
bucket索引中的结点至新表
2.在计算新插入结点键的bucket索引值时, 内部会探测哈希表是否需要扩容(若当前不
在平滑扩容过程中)
1 int dictAdd(dict *d, void *key, void *val)
```

(^2) {
(^3) dictEntry *entry = dictAddRaw(d,key,NULL); // 调用dictAddRaw
(^45) if (!entry) return DICT_ERR;
6 dictSetVal(d, entry, val);
7 return DICT_OK;
8 }
(^109) dictEntry *dictAddRaw(dict *d, void *key, dictEntry **existing)
(^11) {
(^12) long index;
(^13) dictEntry *entry;
(^14) dictht *ht;
(^15) if (dictIsRehashing(d)) _dictRehashStep(d); // 若在平滑扩容过程中, 先步进
迁移一个bucket索引
16
1718 /* Get the index of the new element, or -1 if
(^19) * the element already exists. */
(^2021) // 在计算键在bucket中的索引值时, 内部会检查是否需要扩容
if ((index = _dictKeyIndex(d, key, dictHashKey(d,key), existing)) ==
-1)
22
(^23) return NULL;
(^2425) /* Allocate the memory and store the new entry.
26 * Insert the element in top, with the assumption that in a database
27 * system it is more likely that recently added entries are accessed


```
下面是计算bucket索引值的函数, 内部会探测该哈希表是否需要扩容, 如果需要扩容(结
点数目与bucket数组⻓度比例达到1:1), 就使字典进入平滑扩容过程:
```

28 * more frequently. */
29 ht = dictIsRehashing(d)? &d->ht[1] : &d->ht[0];

(^30) entry = zmalloc(sizeof(*entry));
(^31) entry->next = ht->table[index];
(^32) ht->table[index] = entry;
(^33) ht->used++;
(^3435) /* Set the hash entry fields. */
36 dictSetKey(d, entry, key);
37 return entry;
38 }
39
static long _dictKeyIndex(dict *d, const void *key, uint64_t hash,
dictEntry **existing)
1
(^2) {
(^3) unsigned long idx, table;
(^4) dictEntry *he;
5 if (existing) *existing = NULL;
67 /* Expand the hash table if needed */
if (_dictExpandIfNeeded(d) == DICT_ERR) // 探测是否需要扩容, 如果需要, 则
开始扩容
8
(^9) return -1;
(^10) for (table = 0; table <= 1; table++) {
(^11) idx = hash & d->ht[table].sizemask;
(^12) /* Search if this slot does not already contain the given key */
(^13) he = d->ht[table].table[idx];
14 while(he) {
15 if (key==he->key || dictCompareKeys(d, key, he->key)) {
16 if (existing) *existing = he;
(^17) return -1;
(^18) }
(^19) he = he->next;
(^20) }
(^21) if (!dictIsRehashing(d)) break;
22 }
23 return idx;
24 }
(^2526) /* Expand the hash table if needed */
(^27) static int _dictExpandIfNeeded(dict *d)
(^28) {
(^29) /* Incremental rehashing already in progress. Return. */
if (dictIsRehashing(d)) return DICT_OK; // 如果正在扩容过程中, 则什么也不
做
30
3132 /* If the hash table is empty expand it to the initial size. */
33 // 若字典中本无元素, 则初始化字典, 初始化时的bucket数组长度为 4
34 if (d->ht[0].size == 0) return dictExpand(d, DICT_HT_INITIAL_SIZE);
(^3536) /* If we reached the 1:1 ratio, and we are allowed to resize the hash
(^37) * table (global setting) or we should avoid it but the ratio between


#### 下面是平滑扩容的实现:

38 * elements/buckets is over the "safe" threshold, we resize doubling
39 * the number of buckets. */
// 若字典中元素的个数与bucket数组长度比值大于1:1时, 则调用dictExpand进入平滑
扩容状态

40

(^41) if (d->ht[0].used >= d->ht[0].size &&
(^42) (dict_can_resize ||
(^43) d->ht[0].used/d->ht[0].size > dict_force_resize_ratio))
44 {
45 return dictExpand(d, d->ht[0].used*2);
46 }
(^47) return DICT_OK;
(^48) }
(^4950) int dictExpand(dict *d, unsigned long size)
(^51) {
(^52) dictht n; /* the new hash table */ // 新建一个dictht结构
(^53) unsigned long realsize = _dictNextPower(size);
5455 /* the size is invalid if it is smaller than the number of
56 * elements already inside the hash table */
57 if (dictIsRehashing(d) || d->ht[0].used > size)
(^58) return DICT_ERR;
(^5960) /* Rehashing to the same table size is not useful. */
(^61) if (realsize == d->ht[0].size) return DICT_ERR;
(^6263) /* Allocate the new hash table and initialize all pointers to NULL */
(^64) n.size = realsize;
65 n.sizemask = realsize-1;
n.table = zcalloc(realsize*sizeof(dictEntry*));// 初始化dictht下的
table, 即bucket数组
66
(^67) n.used = 0;
(^6869) /* Is this the first initialization? If so it's not really a rehashing
(^70) * we just set the first hash table so that it can accept keys. */
(^71) // 若是新字典初始化, 直接把dictht结构挂在ht[0]中
(^72) if (d->ht[0].table == NULL) {
(^73) d->ht[0] = n;
74 return DICT_OK;
75 }
76 // 否则, 把新dictht结构挂在ht[1]中, 并开启平滑扩容(置rehashidx为0, 字典处于
非扩容状态时, 该字段值为-1)
77
(^78) /* Prepare a second hash table for incremental rehashing */
(^79) d->ht[1] = n;
(^80) d->rehashidx = 0;
(^81) return DICT_OK;
82 }
1 static void _dictRehashStep(dict *d) {
2 // 若字典上还运行着安全迭代器, 则不迁移结点
3 // 否则每次迁移一个旧bucket索引上的所有结点
(^4) if (d->iterators == 0) dictRehash(d,1);
(^5) }


#### 总结:

```
67 int dictRehash(dict *d, int n) {
8 int empty_visits = n*10; /* Max number of empty buckets to visit. */
```

(^9) if (!dictIsRehashing(d)) return 0;
(^1011) while(n-- && d->ht[0].used != 0) {
(^12) dictEntry *de, *nextde;
(^13) /* Note that rehashidx can't overflow as we are sure there are
more
14
15 * elements because ht[0].used != 0 */
16 assert(d->ht[0].size > (unsigned long)d->rehashidx);
17 // 在旧bucket中, 找到下一个非空的索引位
(^18) while(d->ht[0].table[d->rehashidx] == NULL) {
(^19) d->rehashidx++;
(^20) if (--empty_visits == 0) return 1;
(^21) }
(^22) // 取出该索引位上的结点链表
(^23) de = d->ht[0].table[d->rehashidx];
/* Move all the keys in this bucket from the old to the new hash
HT */
24
25 // 把所有结点迁移到新bucket中去
(^26) while(de) {
(^27) uint64_t h;
(^2829) nextde = de->next;
(^30) /* Get the index in the new hash table */
(^31) h = dictHashKey(d, de->key) & d->ht[1].sizemask;
32 de->next = d->ht[1].table[h];
33 d->ht[1].table[h] = de;
34 d->ht[0].used--;
(^35) d->ht[1].used++;
(^36) de = nextde;
(^37) }
(^38) d->ht[0].table[d->rehashidx] = NULL;
(^39) d->rehashidx++;
(^40) }
4142 /* Check if we already rehashed the whole table... */
43 // 检查是否旧表中的所有结点都被迁移到了新表
44 // 如果是, 则置先释放原旧bucket数组, 再置ht[1]为ht[0]
(^45) // 最后再置rehashidx=-1, 以示字典不处于平滑扩容状态
(^46) if (d->ht[0].used == 0) {
(^47) zfree(d->ht[0].table);
(^48) d->ht[0] = d->ht[1];
(^49) _dictReset(&d->ht[1]);
50 d->rehashidx = -1;
51 return 0;
52 }
(^5354) /* More to rehash... */
(^55) return 1;
(^56) }


#### 字典的实现很复杂, 主要是实现了平滑扩容逻辑 用户数据均是以指针形式间接由

```
dictEntry结构持有, 故在平滑扩容过程中, 不涉及用户数据的拷⻉ 有安全迭代器可用, 安
全迭代器保证, 在迭代起始时, 字典中的所有结点, 都会被迭代到, 即使在迭代过程中对字
典有插入操作 字典内部使用的默认散列函数其实也非常有讲究, 不过限于篇幅, 这里不展
开讲. 并且字典的实现给了使用者非常大的灵活性(dictType结构与dict.privdata字段),
对于一些特定场合使用的键数据, 用户可以自行选择更高效更特定化的散列函数
```

```
2.4 zskiplist
zskiplist是Redis实现的一种特殊的跳跃表. 跳跃表是一种基于线性表实现简单的搜索结
构, 其最大的特点就是: 实现简单, 性能能逼近各种搜索树结构. 血统纯正的跳跃表的介绍
在维基百科中即可查阅. 在Redis中, 在原版跳跃表的基础上, 进行了一些小改动, 即是现
在要介绍的zskiplist结构.
其定义在src/server.h中, 如下:
```

#### 其内存布局如下图:

```
1 /* ZSETs use a specialized version of Skiplists */
2 typedef struct zskiplistNode {
```

(^3) sds ele;
(^4) double score;
(^5) struct zskiplistNode *backward;
(^6) struct zskiplistLevel {
(^7) struct zskiplistNode *forward;
(^8) unsigned int span;
9 } level[];
10 } zskiplistNode;
(^1112) typedef struct zskiplist {
(^13) struct zskiplistNode *header, *tail;
(^14) unsigned long length;
(^15) int level;
(^16) } zskiplist;


```
zskiplist的核心设计要点为:
1.头结点不持有任何数据, 且其level[]的⻓度为3 2
2.每个结点, 除了持有数据的ele字段, 还有一个字段score, 其标示着结点的得分, 结点
之间凭借得分来判断先后顺序, 跳跃表中的结点按结点的得分升序排列.
3.每个结点持有一个backward指针, 这是原版跳跃表中所没有的. 该指针指向结点的前
一个紧邻结点.
4.每个结点中最多持有32个zskiplistLevel结构. 实际数量在结点创建时, 按幂次定律随
机生成(不超过32). 每个zskiplistLevel中有两个字段.
5.forward字段指向比自己得分高的某个结点(不一定是紧邻的), 并且, 若当前
zskiplistLevel实例在level[]中的索引为X, 则其forward字段指向的结点, 其level[]字
段的容量至少是X+1. 这也是上图中, 为什么forward指针总是画的水平的原因.
6.span字段代表forward字段指向的结点, 距离当前结点的距离. 紧邻的两个结点之间
的距离定义为1.
7.zskiplist中持有字段level, 用以记录所有结点(除过头结点外), level[]数组最⻓的⻓度.
跳跃表主要用于, 在给定一个分值的情况下, 查找与该分值最接近的结点. 搜索时, 伪代码
如下:
1 int level = zskiplist->level - 1;
```

(^2) zskiplistNode p = zskiplist->head;
(^34) while(1 && p)
(^5) {
(^6) zskiplistNode q = (p->level)[level]->forward:
(^7) if(q->score > 分值)
8 {
9 if(level > 0)
10 {
(^11) level--;
(^12) }


#### 跳跃表的实现比较简单, 最复杂的操作即是插入与删除结点, 需要仔细处理邻近结点的所

```
有level[]中的所有zskiplistLevel结点中的forward与span的值的变更.
另外, 关于新创建的结点, 其 level[] 数组⻓度的随机算法, 在接口zslInsert的实现中, 核
心代码片断如下:
```

13 else
14 {

(^15) return :
(^16) q为整个跳跃表中, 分值大于指定分值的第一个结点
(^17) q->backward为整个跳跃表中, 分值小于或等于指定分值的最后一个结点
(^18) }
(^19) }
20 else
21 {
22 p = q;
(^23) }
(^24) }
1 zskiplistNode *zslInsert(zskiplist *zsl, double score, sds ele) {
2 //...
34 level = zslRandomLevel(); // 随机生成新结点的, level[]数组的长度
5 if (level > zsl->level) {
(^6) // 若生成的新结点的level[]数组的长度比当前表中所有结点的level[]的长度都大
(^7) // 那么头结点中需要新增几个指向该结点的指针
(^8) // 并刷新ziplist中的level字段
(^9) for (i = zsl->level; i < level; i++) {
(^10) rank[i] = 0;
11 update[i] = zsl->header;
12 update[i]->level[i].span = zsl->length;
13 }
(^14) zsl->level = level;
(^15) }
(^16) x = zslCreateNode(level,score,ele); // 创建新结点
(^17) //... 执行插入操作
(^18) }
(^1920) // 按幂次定律生成小于 32 的随机数的函数
21 // 宏 ZSKIPLIST_MAXLEVEL 的定义为32, 宏 ZSKIPLIST_P 被设定为 0.25
22 // 即
23 // level == 1的概率为 75%
(^24) // level == 2的概率为 75% * 25%
(^25) // level == 3的概率为 75% * 25% * 25%
(^26) // ...
(^27) // level == 31的概率为 0.75 * 0.25^30
(^28) // 而
29 // level == 32的概率为 0.75 * sum(i = 31 ~ +INF){ 0.25^i }
30 int zslRandomLevel(void) {
31 int level = 1;
(^32) while ((random()&0xFFFF) < (ZSKIPLIST_P * 0xFFFF))
(^33) level += 1;


```
2.5 intset
这是一个用于存储在序的整数的数据结构, 也底层数据结构中最简单的一个, 其定义与实
现在src/intest.h与src/intset.c中, 关键定义如下:
```

```
inset结构中的encoding的取值有三个, 分别是宏INTSET_ENC_INT16,
INTSET_ENC_INT32, INTSET_ENC_INT64. length代表其中存储的整数的个数,
contents指向实际存储数值的连续内存区域. 其内存布局如下图所示:
```

```
 intset中各
```

```
字段, 包括contents中存储的数值, 都是以主机序(小端字节序)存储的. 这意味着
Redis若运行在PPC这样的大端字节序的机器上时, 存取数据都会有额外的字节序转
换开销
 当encoding == INTSET_ENC_INT16时, contents中以int16_t的形式存储着数值.
类似的, 当encoding == INTSET_ENC_INT32时, contents中以int32_t的形式存储
着数值.
 但凡有一个数值元素的值超过了int32_t的取值范围, 整个intset都要进行升级, 即所
有的数值都需要以int64_t的形式存储. 显然升级的开销是很大的.
 intset中的数值是以升序排列存储的, 插入与删除的复杂度均为O(n). 查找使用二分
法, 复杂度为O(log_2(n))
 intset的代码实现中, 不预留空间, 即每一次插入操作都会调用zrealloc接口重新分配
内存. 每一次删除也会调用zrealloc接口缩减占用的内存. 省是省了, 但内存操作的时
间开销上升了.
 intset的编码方式一经升级, 不会再降级.
总之, intset适合于如下数据的存储:
 所有数据都位于一个稳定的取值范围中. 比如均位于int16_t或int32_t的取值范围中
 数据稳定, 插入删除操作不频繁. 能接受O(lgn)级别的查找开销
```

```
2.6 ziplist
```

34 return (level<ZSKIPLIST_MAXLEVEL)? level : ZSKIPLIST_MAXLEVEL;
35 }

```
1 typedef struct intset {
```

(^2) uint32_t encoding;
(^3) uint32_t length;
(^4) int8_t contents[];
5 } intset;
67 #define INTSET_ENC_INT16 (sizeof(int16_t))
8 #define INTSET_ENC_INT32 (sizeof(int32_t))
(^9) #define INTSET_ENC_INT64 (sizeof(int64_t))


ziplist是Redis底层数据结构中, 最苟的一个结构. 它的设计宗旨就是: 省内存, 从牙缝里
省内存. 设计思路和TLV一致, 但为了从牙缝里节省内存, 做了很多额外工作.
ziplist的内存布局与intset一样: 就是一块连续的内存空间. 但区域划分比较复杂, 概览如
下图:

 和intset一样, ziplist中的所有值都是以小端序存储的
 zlbytes字段的类型是uint32_t, 这个字段中存储的是整个ziplist所占用的内存的字节
数
 zltail字段的类型是uint32_t, 它指的是ziplist中最后一个entry的偏移量. 用于快速定
位最后一个entry, 以快速完成pop等操作
 zllen字段的类型是uint16_t, 它指的是整个ziplit中entry的数量. 这个值只占16位, 所
以蛋疼的地方就来了: 如果ziplist中entry的数目小于65535, 那么该字段中存储的就
是实际entry的值. 若等于或超过65535, 那么该字段的值固定为65535, 但实际数量
需要一个个entry的去遍历所有entry才能得到.
 zlend是一个终止字节, 其值为全F, 即0xff. ziplist保证任何情况下, 一个entry的首字
节都不会是2 55
在画图展示entry的内存布局之前, 先讲一下entry中都存储了哪些信息:
 每个entry中存储了它前一个entry所占用的字节数. 这样支持ziplist反向遍历.
 每个entry用单独的一块区域, 存储着当前结点的类型: 所谓的类型, 包括当前结点存
储的数据是什么(二进制, 还是数值), 如何编码(如果是数值, 数值如何存储, 如果是二
进制数据, 二进制数据的⻓度)
 最后就是真实的数据了
entry的内存布局如下所示:

prevlen即是"前一个entry所占用的字节数", 它本身是一个变⻓字段, 规约如下:
 若前一个entry占用的字节数小于 254, 则prevlen字段占一字节
 若前一个entry占用的字节数等于或大于 254, 则prevlen字段占五字节: 第一个字节
值为 254, 即0xfe, 另外四个字节, 以uint32_t存储着值.
encoding 字段的规约就复杂了许多
 若数据是二进制数据, 且二进制数据⻓度小于64字节(不包括64), 那么encoding占一
字节. 在这一字节中, 高两位值固定为0, 低六位值以无符号整数的形式存储着二进制
数据的⻓度. 即 00xxxxxx, 其中低六位bitxxxxxx是用二进制保存的数据⻓度.
 若数据是二进制数据, 且二进制数据⻓度大于或等于64字节, 但小于16384(不包括
16384)字节, 那么encoding占用两个字节. 在这两个字节16位中, 第一个字节的高两
位固定为01, 剩余的14个位, 以小端序无符号整数的形式存储着二进制数据的⻓度,
即 01xxxxxx, yyyyyyyy, 其中yyyyyyyy是高八位, xxxxxx是低六位.


####  若数据是二进制数据, 且二进制数据的⻓度大于或等于16384字节, 但小于2^32-1字

```
节, 则encoding占用五个字节. 第一个字节是固定值10000000, 剩余四个字节, 按小
端序uint32_t的形式存储着二进制数据的⻓度. 这也是ziplist能存储的二进制数据的
最大⻓度, 超过2^32-1字节的二进制数据, ziplist无法存储.
 若数据是整数值, 则encoding和data的规约如下:
 首先, 所有存储数值的entry, 其encoding都仅占用一个字节. 并且最高两位均是11
 若数值取值范围位于[0, 12]中, 则encoding和data挤在同一个字节中. 即为1111
0001~1111 1101, 高四位是固定值, 低四位的值从0001至1101, 分别代表 0 ~ 12这十
五个数值
 若数值取值范围位于[-128, -1] [13, 127]中, 则encoding == 0b 1111 1110. 数值存
储在紧邻的下一个字节, 以int8_t形式编码
 若数值取值范围位于[-32768, -129] [128, 32767]中, 则encoding == 0b 1100
```

0. 数值存储在紧邻的后两个字节中, 以小端序int16_t形式编码
    若数值取值范围位于[-8388608, -32769] [32768, 8388607]中, 则encoding == 0b
   1111 0000. 数值存储在紧邻的后三个字节中, 以小端序存储, 占用三个字节.
    若数值取值范围位于[-2^31, -8388609] [8388608, 2^31 - 1]中, 则encoding == 0b
   1101 0000. 数值存储在紧邻的后四个字节中, 以小端序int32_t形式编码
    若数值取值均不在上述范围, 但位于int64_t所能表达的范围内, 则encoding == 0b
   1110 0000, 数值存储在紧邻的后八个字节中, 以小端序int64_t形式编码

```
在大规模数值存储中, ziplist几乎不浪费内存空间, 其苟的程序到达了字节级别, 甚至对于
[0, 12]区间的数值, 连data里的那一个字节也要省下来. 显然, ziplist是一种特别节省内
存的数据结构, 但它的缺点也十分明显:
 和intset一样, ziplist也不预留内存空间, 并且在移除结点后, 也是立即缩容, 这代表每
次写操作都会进行内存分配操作.
 ziplist最蛋疼的一个问题是: 结点如果扩容, 导致结点占用的内存增⻓, 并且超过2 54
字节的话, 可能会导致链式反应: 其后一个结点的entry.prevlen需要从一字节扩容至
五字节. 最坏情况下, 第一个结点的扩容, 会导致整个ziplist表中的后续所有结点的
entry.prevlen字段扩容. 虽然这个内存重分配的操作依然只会发生一次, 但代码中的
时间复杂度是o(N)级别, 因为链式扩容只能一步一步的计算. 但这种情况的概率十分
的小, 一般情况下链式扩容能连锁反映五六次就很不幸了. 之所以说这是一个蛋疼问
题, 是因为, 这样的坏场景下, 其实时间复杂度并不高: 依次计算每个entry新的空间占
用, 也就是o(N), 总体占用计算出来后, 只执行一次内存重分配, 与对应的memmove操
作, 就可以了. 蛋疼说的是: 代码特别难写, 难读. 下面放一段处理插入结点时处理链式
反应的代码片断, 大家自行感受一下:
unsigned char *__ziplistInsert(unsigned char *zl, unsigned char *p,
unsigned char *s, unsigned int slen) {
```

1

2 size_t curlen = intrev32ifbe(ZIPLIST_BYTES(zl)), reqlen;

(^3) unsigned int prevlensize, prevlen = 0;
(^4) size_t offset;
(^5) int nextdiff = 0;
(^6) unsigned char encoding = 0;
long long value = 123456789; /* initialized to avoid warning. Using a
value
7
8 that is easy to see if for some reason
9 we use it uninitialized. */


10 zlentry tail;
1112 /* Find out prevlen for the entry that is inserted. */

(^13) if (p[0] != ZIP_END) {
(^14) ZIP_DECODE_PREVLEN(p, prevlensize, prevlen);
(^15) } else {
(^16) unsigned char *ptail = ZIPLIST_ENTRY_TAIL(zl);
(^17) if (ptail[0] != ZIP_END) {
18 prevlen = zipRawEntryLength(ptail);
19 }
20 }
(^2122) /* See if the entry can be encoded */
(^23) if (zipTryEncoding(s,slen,&value,&encoding)) {
(^24) /* 'encoding' is set to the appropriate integer encoding */
(^25) reqlen = zipIntSize(encoding);
(^26) } else {
/* 'encoding' is untouched, however zipStoreEntryEncoding will use
the
27
28 * string length to figure out how to encode it. */
29 reqlen = slen;
(^30) }
(^31) /* We need space for both the length of the previous entry and
(^32) * the length of the payload. */
(^33) reqlen += zipStorePrevEntryLength(NULL,prevlen);
(^34) reqlen += zipStoreEntryEncoding(NULL,encoding,slen);
3536 /* When the insert position is not equal to the tail, we need to
37 * make sure that the next entry can hold this entry's length in
38 * its prevlen field. */
(^39) int forcelarge = 0;
(^40) nextdiff = (p[0] != ZIP_END)? zipPrevLenByteDiff(p,reqlen) : 0;
(^41) if (nextdiff == -4 && reqlen < 4) {
(^42) nextdiff = 0;
(^43) forcelarge = 1;
(^44) }
4546 /* Store offset because a realloc may change the address of zl. */
47 offset = p-zl;
48 zl = ziplistResize(zl,curlen+reqlen+nextdiff);
(^49) p = zl+offset;
(^5051) /* Apply memory move when necessary and update tail offset. */
(^52) if (p[0] != ZIP_END) {
(^53) /* Subtract one because of the ZIP_END bytes */
(^54) memmove(p+reqlen,p-nextdiff,curlen-offset-1+nextdiff);
5556 /* Encode this entry's raw length in the next entry. */
57 if (forcelarge)
58 zipStorePrevEntryLengthLarge(p+reqlen,reqlen);
(^59) else
(^60) zipStorePrevEntryLength(p+reqlen,reqlen);
(^6162) /* Update offset for tail */
(^63) ZIPLIST_TAIL_OFFSET(zl) =
(^64) intrev32ifbe(intrev32ifbe(ZIPLIST_TAIL_OFFSET(zl))+reqlen);


```
6566 /* When the tail contains more than one entry, we need to take
67 * "nextdiff" in account as well. Otherwise, a change in the
```

(^68) * size of prevlen doesn't have an effect on the *tail* offset. */
(^69) zipEntry(p+reqlen, &tail);
(^70) if (p[reqlen+tail.headersize+tail.len] != ZIP_END) {
(^71) ZIPLIST_TAIL_OFFSET(zl) =
intrev32ifbe(intrev32ifbe(ZIPLIST_TAIL_OFFSET(zl))+nextdiff);
72
73 }
74 } else {
(^75) /* This element will be the new tail. */
(^76) ZIPLIST_TAIL_OFFSET(zl) = intrev32ifbe(p-zl);
(^77) }
(^78) /* When nextdiff != 0, the raw length of the next entry has changed,
so
79
(^80) * we need to cascade the update throughout the ziplist */
81 if (nextdiff != 0) {
82 offset = p-zl;
83 zl = __ziplistCascadeUpdate(zl,p+reqlen);
(^84) p = zl+offset;
(^85) }
(^8687) /* Write the entry */
(^88) p += zipStorePrevEntryLength(p,prevlen);
(^89) p += zipStoreEntryEncoding(p,encoding,slen);
90 if (ZIP_IS_STR(encoding)) {
91 memcpy(p,s,slen);
92 } else {
(^93) zipSaveInteger(p,value,encoding);
(^94) }
(^95) ZIPLIST_INCR_LENGTH(zl,1);
(^96) return zl;
(^97) }
(^98) unsigned char *__ziplistCascadeUpdate(unsigned char *zl, unsigned char *p)
{
99
100 size_t curlen = intrev32ifbe(ZIPLIST_BYTES(zl)), rawlen, rawlensize;
101 size_t offset, noffset, extra;
(^102) unsigned char *np;
(^103) zlentry cur, next;
(^104105) while (p[0] != ZIP_END) {
(^106) zipEntry(p, &cur);
(^107) rawlen = cur.headersize + cur.len;
108 rawlensize = zipStorePrevEntryLength(NULL,rawlen);
109110 /* Abort if there is no next entry. */
111 if (p[rawlen] == ZIP_END) break;
(^112) zipEntry(p+rawlen, &next);
(^113114) /* Abort when "prevlen" has not changed. */
(^115) if (next.prevrawlen == rawlen) break;
(^116117) if (next.prevrawlensize < rawlensize) {
(^118) /* The "prevlen" field of "next" needs more bytes to hold


#### 这种代码的特点就是: 最好由作者去维护, 最好一次性写对. 因为读起来真的费劲, 改起来

#### 也很费劲.

```
2.7 quicklist
如果说ziplist是整个Redis中为了节省内存, 而写的最苟的数据结构, 那么称quicklist就是
在最苟的基础上, 再苟了一层. 这个结构是Redis在3.2版本后新加的, 在3.2版本之前, 我
们可以讲, dict是最复杂的底层数据结构, ziplist是最苟的底层数据结构. 在3.2版本之后,
这两个记录被双双刷新了.
```

119 * the raw length of "cur". */
120 offset = p-zl;

(^121) extra = rawlensize-next.prevrawlensize;
(^122) zl = ziplistResize(zl,curlen+extra);
(^123) p = zl+offset;
(^124125) /* Current pointer and offset for next element. */
(^126) np = p+rawlen;
127 noffset = np-zl;
128 /* Update tail offset when next element is not the tail
element. */
129
(^130) if ((zl+intrev32ifbe(ZIPLIST_TAIL_OFFSET(zl))) != np) {
(^131) ZIPLIST_TAIL_OFFSET(zl) =
intrev32ifbe(intrev32ifbe(ZIPLIST_TAIL_OFFSET(zl))+extra);
132
(^133) }
(^134135) /* Move the tail to the back. */
136 memmove(np+rawlensize,
137 np+next.prevrawlensize,
138 curlen-noffset-next.prevrawlensize-1);
(^139) zipStorePrevEntryLength(np,rawlen);
(^140141) /* Advance the cursor */
(^142) p += rawlen;
(^143) curlen += extra;
(^144) } else {
145 if (next.prevrawlensize > rawlensize) {
146 /* This would result in shrinking, which we want to avoid.
147 * So, set "rawlen" in the available bytes. */
(^148) zipStorePrevEntryLengthLarge(p+rawlen,rawlen);
(^149) } else {
(^150) zipStorePrevEntryLength(p+rawlen,rawlen);
(^151) }
(^152153) /* Stop here, as the raw length of "next" has not changed. */
(^154) break;
155 }
156 }
157 return zl;
(^158) }


```
这是一种, 以ziplist为结点的, 双端链表结构. 宏观上, quicklist是一个链表, 微观上, 链表
中的每个结点都是一个ziplist.
它的定义与实现分别在src/quicklist.h与src/quicklist.c中, 其中关键定义如下:
/* Node, quicklist, and Iterator are the only data structures used
currently. */
```

```
1
```

```
23 /* quicklistNode is a 32 byte struct describing a ziplist for a quicklist.
```

(^4) * We use bit fields keep the quicklistNode at 32 bytes.

* count: 16 bits, max 65536 (max zl bytes is 65k, so max count actually <
  32k).
  5
  (^6) * encoding: 2 bits, RAW=1, LZF=2.
  (^7) * container: 2 bits, NONE=1, ZIPLIST=2.
* recompress: 1 bit, bool, true if node is temporarry decompressed for
  usage.
  8
  9 * attempted_compress: 1 bit, boolean, used for verifying during testing.
* extra: 12 bits, free for future use; pads out the remainder of 32 bits
  */
  10
  (^11) typedef struct quicklistNode {
  (^12) struct quicklistNode *prev;
  (^13) struct quicklistNode *next;
  (^14) unsigned char *zl;
  15 unsigned int sz; /* ziplist size in bytes */
  16 unsigned int count : 16; /* count of items in ziplist */
  17 unsigned int encoding : 2; /* RAW==1 or LZF==2 */
  (^18) unsigned int container : 2; /* NONE==1 or ZIPLIST==2 */
  (^19) unsigned int recompress : 1; /* was this node previous compressed? */
  unsigned int attempted_compress : 1; /* node can't compress; too small
  */
  20
  (^21) unsigned int extra : 10; /* more bits to steal for future usage */
  (^22) } quicklistNode;
  23 /* quicklistLZF is a 4+N byte struct holding 'sz' followed by
  'compressed'.
  24
  25 * 'sz' is byte length of 'compressed' field.
  (^26) * 'compressed' is LZF data with total (compressed) length 'sz'
  (^27) * NOTE: uncompressed length is stored in quicklistNode->sz.
* When quicklistNode->zl is compressed, node->zl points to a quicklistLZF
  */
  28
  (^29) typedef struct quicklistLZF {
  30 unsigned int sz; /* LZF size in bytes*/
  31 char compressed[];
  32 } quicklistLZF;
  (^33) /* quicklist is a 40 byte struct (on 64-bit systems) describing a
  quicklist.
  34
  (^35) * 'count' is the number of total entries.
  (^36) * 'len' is the number of quicklist nodes.
  (^37) * 'compress' is: -1 if compression disabled, otherwise it's the number
* of quicklistNodes to leave uncompressed at ends of
  quicklist.
  38
  39 * 'fill' is the user-requested (or default) fill factor. */


#### 这里定义了五个结构体:

```
 quicklistNode, 宏观上, quicklist是一个链表, 这个结构描述的就是链表中的结点. 它
通过zl字段持有底层的ziplist. 简单来讲, 它描述了一个ziplist实例
 quicklistLZF, ziplist是一段连续的内存, 用LZ4算法压缩后, 就可以包装成一个
quicklistLZF结构. 是否压缩quicklist中的每个ziplist实例是一个可配置项. 若这个配
置项是开启的, 那么quicklistNode.zl字段指向的就不是一个ziplist实例, 而是一个压
缩后的quicklistLZF实例
 quicklist. 这就是一个双链表的定义. head, tail分别指向头尾指针. len代表链表中的
结点. count指的是整个quicklist中的所有ziplist中的entry的数目. fill字段影响着每
个链表结点中ziplist的最大占用空间, compress影响着是否要对每个ziplist以LZ4算
法进行进一步压缩以更节省内存空间.
 quicklistIter是一个迭代器
 quicklistEntry是对ziplist中的entry概念的封装. quicklist作为一个封装良好的数据
结构, 不希望使用者感知到其内部的实现, 所以需要把ziplist.entry的概念重新包装一
下.
quicklist的内存布局图如下所示:
```

40 typedef struct quicklist {
41 quicklistNode *head;

(^42) quicklistNode *tail;
unsigned long count; /* total count of all entries in all
ziplists */
43
(^44) unsigned long len; /* number of quicklistNodes */
(^45) int fill : 16; /* fill factor for individual nodes */
unsigned int compress : 16; /* depth of end nodes not to
compress;0=off */
46
47 } quicklist;
(^4849) typedef struct quicklistIter {
(^50) const quicklist *quicklist;
(^51) quicklistNode *current;
(^52) unsigned char *zi;
(^53) long offset; /* offset in current ziplist */
(^54) int direction;
55 } quicklistIter;
5657 typedef struct quicklistEntry {
58 const quicklist *quicklist;
(^59) quicklistNode *node;
(^60) unsigned char *zi;
(^61) unsigned char *value;
(^62) long long longval;
(^63) unsigned int sz;
64 int offset;
65 } quicklistEntry;


下面是有关quicklist的更多额外信息:
quicklist.fill的值影响着每个链表结点中, ziplist的⻓度.
1.当数值为负数时, 代表以字节数限制单个ziplist的最大⻓度. 具体为:
a.-1 不超过4kb
b.-2 不超过 8kb
c.-3 不超过 16kb
d.-4 不超过 32kb
e.-5 不超过 64kb
f. 当数值为正数时, 代表以entry数目限制单个ziplist的⻓度. 值即为数目. 由于该字段仅占
16位, 所以以entry数目限制ziplist的容量时, 最大值为2^15个
2.quicklist.compress的值影响着quicklistNode.zl字段指向的是原生的ziplist, 还是经
过压缩包装后的quicklistLZF
a.0 表示不压缩, zl字段直接指向ziplist
b.1 表示quicklist的链表头尾结点不压缩, 其余结点的zl字段指向的是经过压缩后的
quicklistLZF
c.2 表示quicklist的链表头两个, 与末两个结点不压缩, 其余结点的zl字段指向的是经过压
缩后的quicklistLZF
d.以此类推, 最大值为2^16
3.quicklistNode.encoding字段, 以指示本链表结点所持有的ziplist是否经过了压缩. 1
代表未压缩, 持有的是原生的ziplist, 2代表压缩过
4.quicklistNode.container字段指示的是每个链表结点所持有的数据类型是什么. 默认
的实现是ziplist, 对应的该字段的值是2, 目前Redis没有提供其它实现. 所以实际上,
该字段的值恒为2
5.quicklistNode.recompress字段指示的是当前结点所持有的ziplist是否经过了解压.
如果该字段为1即代表之前被解压过, 且需要在下一次操作时重新压缩.
quicklist的具体实现代码篇幅很⻓, 这里就不贴代码片断了, 从内存布局上也能看出来,
由于每个结点持有的ziplist是有上限⻓度的, 所以在与操作时要考虑的分支情况比较多.
想想都蛋疼.
quicklist有自己的优点, 也有缺点, 对于使用者来说, 其使用体验类似于线性数据结构,
list作为最传统的双链表, 结点通过指针持有数据, 指针字段会耗费大量内存. ziplist解决
了耗费内存这个问题. 但引入了新的问题: 每次写操作整个ziplist的内存都需要重分配.


quicklist在两者之间做了一个平衡. 并且使用者可以通过自定义quicklist.fill, 根据实际业
务情况, 经验主义调参.

2.8 zipmap

dict作为字典结构, 优点很多, 扩展性强悍, 支持平滑扩容等等, 但对于字典中的键值均为
二进制数据, 且⻓度都很小时, dict的中的一坨指针会浪费不少内存, 因此Redis又实现了
一个轻量级的字典, 即为zipmap.
zipmap适合使用的场合是:
 键值对量不大, 单个键, 单个值⻓度小
 键值均是二进制数据, 而不是复合结构或复杂结构. dict支持各种嵌套, 字典本身并不
持有数据, 而仅持有数据的指针. 但zipmap是直接持有数据的.
zipmap的定义与实现在src/zipmap.h与src/zipmap.c两个文件中, 其定义与实现均未定
义任何struct结构体, 因为zipmap的内存布局就是一块连续的内存空间. 其内存布局如下
所示:

 zipmap起始的第一个字节存储的是zipmap中键值对的个数. 如果键值对的个数大于
254的话, 那么这个字节的值就是固定值254, 真实的键值对个数需要遍历才能获得.
 zipmap的最后一个字节是固定值0xFF
 zipmap中的每一个键值对, 称为一个entry, 其内存占用如上图, 分别六部分:
 len_of_key, 一字节或五字节. 存储的是键的二进制⻓度. 如果⻓度小于254, 则用1字节
存储, 否则用五个字节存储, 第一个字节的值固定为0xFE, 后四个字节以小端序uint32_t
类型存储着键的二进制⻓度.
 key_data为键的数据
 len_of_val, 一字节或五字节, 存储的是值的二进制⻓度. 编码方式同len_of_key
 len_of_free, 固定值1字节, 存储的是entry中未使用的空间的字节数. 未使用的空间即为
图中的free, 它一般是由于键值对中的值被替换发生的. 比如, 键值对hello <-> word被
修改为hello <-> w后, 就空了四个字节的闲置空间
 val_data, 为值的数据
 free, 为闲置空间. 由于len_of_free的值最大只能是254, 所以如果值的变更导致闲置空
间大于254的话, zipmap就会回收内存空间.

### Redis中内存淘汰算法实现

Redis的maxmemory 支持的内存淘汰机制使得其成为一种有效的缓存方案，成为
memcached的有效替代方案。
当内存达到 maxmemory后，Redis会按照 maxmemory-policy 启动淘汰策略。
Redis 3.0中已有淘汰机制：
 noeviction


 allkeys-lru
 volatile-lru
 allkeys-random
 volatile-random
 volatile-ttl

其中LRU(less recently used)经典淘汰算法在Redis实现中有一定优化设计，来保证内
存占用与实际效果的平衡，这也体现了工程应用是空间与时间的平衡性。

```
PS：值得注意的，在主从复制模式Replication下，从节点达到maxmemory时不会有
任何异常日志信息，但现象为增量数据无法同步至从节点。
```

### Redis 3.0中近似LRU算法

Redis中LRU是近似LRU实现，并不能取出理想LRU理论中最佳淘汰Key，而是通过从小
部分采样后的样本中淘汰局部LRU键。
Redis 3.0中近似LRU算法通过增加待淘汰元素池的方式进一步优化，最终实现与精确
LRU非常接近的表现。

```
精确LRU会占用较大内存记录历史状态，而近似LRU则用较小内存支出实现近似效
果。
```

以下是理论LRU和近似LRU的效果对比：

```
maxmemory-
policy
```

#### 含义 特性

```
noeviction 不淘汰 内存超限后写命令会返回错误(如OOM, del命令
除外)
allkeys-lru 所有key的LRU
机制 在
```

```
所有key中按照最近最少使用LRU原则剔除key，
释放空间
volatile-lru 易失key的LRU 仅以设置过期时间key范围内的LRU(如均为设置
过期时间，则不会淘汰)
allkeys-random 所有key随机淘
汰
```

#### 一视同仁，随机

```
volatile-random 易失Key的随机 仅设置过期时间key范围内的随机
```

```
volatile-ttl 易失key的TTL
淘汰
```

```
按最小TTL的key优先淘汰
```

####  按时间顺序接入不同键，此时最早写入也就是最佳淘汰键

####  浅灰色区域：被淘汰的键

####  灰色区域：未被淘汰的键

####  绿色区域：新增写入的键

#### 总结图中展示规律，

```
 图1Theoretical LRU符合预期：最早写入键逐步被淘汰
 图2Approx LRU Redis 3.0 10 samples：Redis 3.0中近似LRU算法(采样值为10)
 图3Approx LRU Redis 2.8 5 samples：Redis 2.8中近似LRU算法(采样值为5)
 图4Approx LRU Redis 3.0 5 samples：Redis 3.0中近似LRU算法(采样值为5)
结论：
 通过图4和图3对比：得出相同采样值下，3.0比2.8的LRU淘汰机制更接近理论LRU
 通过图4和图2对比：得出增加采样值，在3.0中将进一步改善LRU淘汰效果逼近理论
LRU
 对比图2和图1：在3.0中采样值为10时，效果非常接近理论LRU
采样值设置通过maxmemory-samples指定，可通过CONFIG SET maxmemory-
samples 动态设置，也可启动配置中指定maxmemory-samples
源码解析
1 int freeMemoryIfNeeded(void){
```

(^2) while (mem_freed < mem_tofree) {
(^3) if (server.maxmemory_policy == REDIS_MAXMEMORY_NO_EVICTION)
return REDIS_ERR; /* We need to free memory, but policy forbids.
*/
4
56 if (server.maxmemory_policy == REDIS_MAXMEMORY_ALLKEYS_LRU ||
7 server.maxmemory_policy == REDIS_MAXMEMORY_ALLKEYS_RANDOM)
(^8) {......}
(^9) /* volatile-random and allkeys-random policy */
(^10) if (server.maxmemory_policy == REDIS_MAXMEMORY_ALLKEYS_RANDOM ||
server.maxmemory_policy ==
REDIS_MAXMEMORY_VOLATILE_RANDOM)
11
(^12) {......}
13 /* volatile-lru and allkeys-lru policy */
14 else if (server.maxmemory_policy == REDIS_MAXMEMORY_ALLKEYS_LRU ||


15 server.maxmemory_policy == REDIS_MAXMEMORY_VOLATILE_LRU)
16 {

(^17) // 淘汰池函数
(^18) evictionPoolPopulate(dict, db->dict, db->eviction_pool);
(^19) while(bestkey == NULL) {
(^20) evictionPoolPopulate(dict, db->dict, db->eviction_pool);
(^21) // 从后向前逐一淘汰
22 for (k = REDIS_EVICTION_POOL_SIZE-1; k >= 0; k--) {
23 if (pool[k].key == NULL) continue;
24 de = dictFind(dict,pool[k].key); // 定位目标
(^2526) /* Remove the entry from the pool. */
(^27) sdsfree(pool[k].key);
(^28) /* Shift all elements on its right to left. */
(^29) memmove(pool+k,pool+k+1,
(^30) sizeof(pool[0])*(REDIS_EVICTION_POOL_SIZE-k-1));
(^31) /* Clear the element on the right which is empty
32 * since we shifted one position to the left. */
33 pool[REDIS_EVICTION_POOL_SIZE-1].key = NULL;
34 pool[REDIS_EVICTION_POOL_SIZE-1].idle = 0;
(^3536) /* If the key exists, is our pick. Otherwise it is
(^37) * a ghost and we need to try the next element. */
(^38) if (de) {
(^39) bestkey = dictGetKey(de); // 确定删除键
(^40) break;
41 } else {
42 /* Ghost... */
43 continue;
(^44) }
(^45) }
(^46) }
(^47) }
(^48) /* volatile-ttl */
else if (server.maxmemory_policy == EDIS_MAXMEMORY_VOLATILE_TTL)
{......}
49
5051 // 最终选定待删除键bestkey
52 if (bestkey) {
(^53) long long delta;
robj *keyobj = createStringObject(bestkey,sdslenbestkey)); //
目标对象
54
(^55) propagateExpire(db,keyobj);
(^56) latencyStartMonitor(eviction_latency); // 延迟监控开始
57 dbDelete(db,keyobj); // 从db删除对象
58 latencyEndMonitor(eviction_latency);// 延迟监控结束
latencyAddSampleIfNeeded("eviction-del",iction_latency); // 延
迟采样
59
(^60) latencyRemoveNestedEvent(latency,eviction_latency);
(^61) delta -= (long long) zmalloc_used_memory();
(^62) mem_freed += delta; // 释放内存计数
(^63) server.stat_evictedkeys++; // 淘汰key计数，info中可见


### Redis 4.0中新的LFU算法

```
从Redis4.0开始，新增LFU淘汰机制，提供更好缓存命中率。LFU(Least Frequently
Used)通过记录键使用频率来定位最可能淘汰的键。
对比LRU与LFU的差别：
 在LRU中，某个键很少被访问，但在刚刚被访问后其被淘汰概率很低，从而出现这类
异常持续存在的缓存；相对的，其他可能被访问的键会被淘汰
 而LFU中，按访问频次淘汰最少被访问的键
Redis 4.0中新增两种LFU淘汰机制：
 volatile-lfu：设置过期时间的键按LFU淘汰
 allkeys-lfu：所有键按LFU淘汰
LFU使用Morris counters计数器占用少量位数来评估每个对象的访问频率，并随时间更
新计数器。此机制实现与近似LRU中采样类似。但与LRU不同，LFU提供明确参数来指
定计数更新频率。
 lfu-log-factor：0-255之间，饱和因子，值越小代表饱和速度越快
 lfu-decay-time：衰减周期，单位分钟，计数器衰减的分钟数
这两个因子形成一种平衡，通过少量访问 VS 多次访问 的评价标准最终形成对键重要性
的评判。
原文： http://fivezh.github.io/2019/01/10/Redis-LRU-algorithm/
```

### Redis中内存淘汰算法实现

```
Redis的maxmemory 支持的内存淘汰机制使得其成为一种有效的缓存方案，成为
memcached的有效替代方案。
当内存达到 maxmemory后，Redis会按照 maxmemory-policy 启动淘汰策略。
Redis 3.0中已有淘汰机制：
 noeviction
 allkeys-lru
 volatile-lru
 allkeys-random
 volatile-random
 volatile-ttl
```

```
notifyKeyspaceEvent(REDIS_NOTIFY_EVICTED, "evicted", keyobj,
db->id); // 事件通知
```

64

(^65) decrRefCount(keyobj); // 引用计数更新
(^66) keys_freed++;
(^67) // 避免删除较多键导致的主从延迟，在循环内同步
(^68) if (slaves) flushSlavesOutputBuffers();
(^69) }
70 }
71 }


其中LRU(less recently used)经典淘汰算法在Redis实现中有一定优化设计，来保证内
存占用与实际效果的平衡，这也体现了工程应用是空间与时间的平衡性。

```
PS：值得注意的，在主从复制模式Replication下，从节点达到maxmemory时不会有
任何异常日志信息，但现象为增量数据无法同步至从节点。
```



参考链接：

https://www.cnblogs.com/neooelric/p/9621736.html