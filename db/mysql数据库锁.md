## MySQL的InnoDB的锁机制

MySQL的InnoDB引擎下，在锁的级别上一般分为两种：**共享锁（S锁）**、**排他锁（X锁）**

### 共享锁

**`共享锁又称为读锁`**，**是读取操作时创建的锁。其他用户可以并发读取数据，但是一旦某行被加上共享锁，其他事务仍可继续对该行加共享锁；但任何事务若想对该行加排他锁（即执行 `UPDATE`/`DELETE` 等），则必须等待所有共享锁释放**。

> 例如：事务T1对数据A的加上共享锁，其他事务只能对数据A加共享锁，不能加排他锁。而共享锁的事务只能读取数据，不能修改数据。

共享锁的加锁方式如下：

```
-- MySQL8.0之前的推荐写法
SELECTID,NAMEFROM TABLE1 WHEREID >1ANDID <10LOCKINSHAREMODE;

-- MySQL8.0以及之后的推荐写法
SELECTID,NAMEFROM TABLE1 WHEREID >1ANDID <10FORSHARE;
```

在查询语句后面增加`LOCK IN SHARE MODE`，会对查询范围中的每行都加共享锁，这样的数据行还可以被其他事务成功申请共享锁，但是不能被申请排他锁。

### 排他锁

**`排他锁又称写锁`，若事务T1对数据A加上排他锁之后，其他事务则不能再对数据A加任何类型的锁。而且排他锁的事务既可以读数据又可以写数据。**

排他锁的加锁方式：

```
SELECT ID,NAME FROM TABLE1 WHERE ID >1 AND ID <10 FOR UPDATE;
```

在查询语句后面增加`FOR UPDATE`，会对查询命中的每条记录都加排他锁，当没有其他线程对查询结果集中的任何一行使用排他锁时，可以成功申请排他锁，否则会被阻塞。

### 共享锁和排他锁的总结

* 当一行数据获取了排他锁，那么其他事务就不能再对这一行数据添加**共享锁**或者**排他锁**。
* 当一行数据获取了共享锁，那么其他事务依然可以对这一行数据添加**共享锁**，但不能添加**排他锁**。

#### 使用场景

共享锁

* **读-读并发高且不允许“脏读”：例如报表统计时，希望别的事务可以并发读，但禁止任何事务修改这些行。**
* **父子表一致性校验：先对父表主键 FOR SHARE，再读子表，防止父记录在此期间被删。**

排他锁

* 读-改-写（Read-Modify-Write）：

```
SELECT balance FROM account WHERE id = 1 FOR UPDATE;
UPDATE account SET balance = balance - 100 WHERE id = 1;
```

防止并发扣款出现负余额。

* 悲观锁实现“秒杀库存”：先 FOR UPDATE 检查库存 > 0，再 UPDATE 减库存。

```
-- 加排他锁并读取当前库存
SELECT stock FROM merchandise WHERE id = 123 FOR UPDATE;

-- 在应用层判断返回的 stock 值,若 stock > 0，则继续；否则回滚并返回“已售罄”
-- 真正扣库存（MySQL 8 支持原子写法，也可拆两步）
UPDATE merchandise SET stock = stock - 1 WHERE id = 123 AND stock > 0;
```

### 意向锁

**除了S锁（共享锁）和X锁（排他锁）之外，InnoDB还有两种锁，就是IS锁和IX锁，S和X前面的I是`Intention`的意思，即意向锁，IS就是意向共享锁，IX就是意向排他锁。**

在MySQL的InnoDB引擎中，根据锁的不同范围也是有区分的，例如：表级锁、间隙锁、行级锁等。当多个事务同时访问同一个数据时，多个事务同时申请获取锁，那么就有可能导致互相阻塞甚至产生死锁。

例如：

> 事务T1对表Table1中的一行加上了行级锁，此时这行记录就不能被其他事务写了。 事务T2申请对Table1增加了表级锁，若申请成功了，那么就可以修改表中的任意一行记录。 这就跟事务T1发生了冲突。 那么，想要解决这个问题，就需要让事务T2在对Table1增加表级锁的时候，先判断一下是不是有事务增加过行级锁。但是，事务T2总不能逐条判断是否有加锁吧?

![](https://mmbiz.qpic.cn/mmbiz_jpg/51zXejl2tKzSiar6k49nriaHRIfGyNoGKicMoYlVAq3VzeicttcO1deQkVeicOfN5AWxBibjN0ZxSAUzx4BBGX2nDGVw/640?wx_fmt=jpeg&from=appmsg&watermark=1)

因此，为了解决这个问题，MySQL引入了意向锁机制，\*\*`意向锁是数据库管理系统中用于实现锁协议的一种锁机制，主要是用来处理不同粒度锁（如行锁和表锁）之间的并发性问题（而同粒度的锁之间一般是通过互斥来解决并发的）`\*\*。

> * 意向锁不做锁定资源的操作，主要是通知的作用，防止在已经加锁的数据上设置不兼容的锁。
> * 意向锁不是直接由用户请求的，而是MySQL管理的。

当一个事务想获取一个行级锁或表级锁时，MySQL会自动获取相应的表的意向锁。这时，其他事务再请求获取表锁时，就可以先基于这个意向锁来发现是否已经加过锁并且未释放，并根据该锁的类型(意向共享锁/意向排他锁)来判断自己是否可以获取锁。这样可以在不阻塞其他事务的情况下，为当前事务锁定资源。

#### 举个🌰：

![](https://mmbiz.qpic.cn/mmbiz_png/51zXejl2tKzSiar6k49nriaHRIfGyNoGKicvKyTIUEz36HSIP7qESzluzdHXFEoBklwknAHfUVuicNpViaiaYkRHTCjw/640?wx_fmt=png&from=appmsg&watermark=1)

1、事务 T1 执行

```
SELECT * FROM table1 WHERE id=1 FOR UPDATE;
```

InnoDB 先对表 table1 加 IX，再对 id=1 的行加 X 锁。

2、此时事务 T2 想 申请表锁

```
LOCK TABLES table1 WRITE;
```

需等待，因为检测到表 table1 上已有 IX，知道“\*\*`table1`里面某行正被排他锁住\*\*”，于是阻塞。 3、事务T1提交成功，释放table1的IX锁，这个时候T2被唤醒成功获取`table1`的IX锁。

从这个例子中可以看出来，**意向锁是个表级锁，并且会在触发意向锁的事务提交或回滚后释放。**

#### 意向锁有两种：**意向共享锁**、**意向排他锁**

* **意向共享锁（IS）**:表示事务打算在资源上设置共享锁(读锁)。这通常用于表示事务计划读取资源，并不希望在读取时有其他事务设置排它锁。
* **意向排它锁（IX）**:表示事务打算在资源上设置排它锁(写锁)。这表示事务计划修改资源，并不希望有其他事务同时设置共享或排它锁。

这两种意向锁的兼容情况如下：


| 已持有\\ 请求        | IS（意向共享锁） | IX（意向排他锁） | S（表级共享） | X（表级排他） |
| -------------------- | ---------------- | ---------------- | ------------- | ------------- |
| **IS（意向共享锁)**  | 兼容             | 兼容             | 兼容          | 冲突          |
| **IX（意向排他锁）** | 兼容             | 兼容             | 兼容          | 冲突          |
| **S（表级共享）**    | 兼容             | 冲突             | 兼容          | 冲突          |
| **X（表级排他）**    | 冲突             | 冲突             | 冲突          | 冲突          |

### 记录锁（Record Lock）

Record Lock 解释为记录锁，是因为它只对索引记录加锁。 例如：

```
SELECT * FROM users WHERE id = 5 FOR UPDATE;
```

上面这个SQL会对 id = 5这条记录加锁，主要就是为了防止其他事务，对id =5这条记录进行插入、更新、删除等操作。

Record Lock 记录锁可以是\*\*共享锁(S锁)**或**排他锁(X锁)\*\*。

> 有一点需要注意的是，Record Lock锁的不是这行记录，而是锁索引记录。并且Record lock锁且只锁索引! 如果没有索引怎么办?对于这种情况，InnoDB会创建一个隐藏的聚簇索引，并使用这个索引进行记录锁定。 如果我们在一张表中没有定义主键，那么，MySQL会默认选择一个唯一的非空索引作为聚簇索引。如果没有适合的非空唯一索引，则会创建一个隐藏的主键(row\_id)作为聚簇索引。

![](https://mmbiz.qpic.cn/mmbiz_png/51zXejl2tKzSiar6k49nriaHRIfGyNoGKiciaicuEM6fGQ87icFdmUcsA83LRzEL9Efg9wDh3iaibp35iaJRXL1wb8kTqibA/640?wx_fmt=png&from=appmsg&watermark=1)

### Gap Lock（间隙锁）

**Gap Lock，间隙锁，是在索引记录之间的间隙上的锁，或者在第一个索引记录之前、最后一个索引记录之后的间隙上加锁。**

**Gap Lock 锁住的是索引记录之间的开区间，不包含记录本身。**

> Gap也可以理解为InnoDB的索引数据结构中可以插入新值的位置。

![](https://mmbiz.qpic.cn/mmbiz_png/51zXejl2tKzSiar6k49nriaHRIfGyNoGKicvY27bMykfk3NgtqNsIuicg223EfevKoy8HqY6pGD24G8iaOqaNMS5P6Q/640?wx_fmt=png&from=appmsg&watermark=1)

示例：

```
SELECT * FROM users WHERE id = 7 FOR UPDATE;
```

由于索引记录不存在，加Gap Lock：区间(5,10)。这样就能防止其他事务在间隙中插入新记录。

### Next Key Lock （临键锁）

**Next-Key Lock是索引记录上的记录锁和索引记录之前间隙上的间隙锁的组合。**

![](https://mmbiz.qpic.cn/mmbiz_png/51zXejl2tKzSiar6k49nriaHRIfGyNoGKicoF7O1dr2SxBXxyFicvTdDG96gzWruhwQqU09Ik3Kexuntct3uicbAiciaQ/640?wx_fmt=png&from=appmsg&watermark=1)

工作示例：

```
SELECT * FROM users WHERE age = 10 FOR UPDATE;
```

**如果上面这个SQL中users表的age为普通索引，那么执行这个SQL的时候，会添加Next-Key Lock，并锁住（5,10]这个范围，左开右闭。主要是对age索引：(5,10]的Next-Key Lock，对主键索引：id=10的记录锁**。

### 加锁规则与时机

#### 加锁基本原则：

1. 所有锁都是加在索引上的。
2. 加锁的基本单位是 Next-Key Lock。
3. 查询过程中访问到的对象都会加锁。
4. 加锁顺序是从左向右，直到第一个不满足条件的记录。

> 这些锁都是MySQL的InnoDB自动加上的，不需要用户手动操作，但是我们在日常使用数据库的时候要明白加锁的原则。

#### 不同场景的加锁规则：

1. **等值查询 (=)****唯一索引/主键**：Next-Key Lock 退化为 Record Lock**普通索引**：保持 Next-Key Lock，左右间隙都不能优化掉
2. **范围查询 (>、<、>=、<=)****唯一索引/主键**：保持 Next-Key Lock，不进行优化**普通索引**：同样保持 Next-Key Lock
3. **记录不存在的情况**会对扫描到的第一个不满足条件的记录加 Gap Lock

#### 工作场景示例

例如有一个users表字段与现有数据如下表所示


| id | age | name |
| -- | --- | ---- |
| 1  | 5   | Tom  |
| 5  | 10  | Jack |
| 10 | 15  | Rose |
| 15 | 20  | Mike |

##### 场景一：等值查询（唯一索引）

```
SELECT * FROM users WHERE id = 5 FOR UPDATE;
```

加锁情况：

* 由于id是主键（唯一索引），Next-Key Lock退化为Record Lock
* 只在id=5的记录上加X型记录锁

##### 场景二：等值查询（普通索引）

```
SELECT * FROM users WHERE age = 10 FOR UPDATE;
```

加锁情况：

* 对age索引：(5,10]的Next-Key Lock
* 对主键索引：id=5的记录锁
* 因为是普通索引，间隙不能被优化掉

##### 场景三：范围查询

```
SELECT * FROM users WHERE id > 5 AND id <= 15 FOR UPDATE;
```

加锁情况：

* (5,10]的Next-Key Lock
* (10,15]的Next-Key Lock
* (15,20]的Next-Key Lock（会扫描到第一个不满足条件的记录）

##### 场景四：记录不存在的情况

```
SELECT * FROM users WHERE id = 7 FOR UPDATE;
```

加锁情况：**由于记录不存在，加Gap Lock：(5,10)**

#### 不同隔离级别下可加锁的行为

![](https://mmbiz.qpic.cn/mmbiz_png/51zXejl2tKzSiar6k49nriaHRIfGyNoGKic44IYCWB6xEyd2AGvIysialcRz7TPW28Vn7a7HMNPkESAQKuoTjUjJYw/640?wx_fmt=png&from=appmsg&watermark=1)

### 插入意向锁（Insert Intention Lock）

**插入意向锁是一种特殊的间隙锁，表示事务计划在某个间隙插入数据的意图。**

这种锁表明了插入的意图，以这样一种方式，如果多个事务插入到同一索引间隙中但不在间隙内的相同位置插入，则它们不需要相互等待。

假设有索引记录的值为4和7。分别尝试插入值为5和6的不同事务，在获取插入行的独占锁之前，各自用插入意向锁锁定4和7之间的间隙，但由于行不冲突，所以它们不会相互阻塞。但是如果他们的都要插入6，那么就会需要阻塞了。

插入意向锁，是一种优化后的间隙锁，主要目的是**优化并发，允许不冲突的插入**。它与间隙锁配合使用，既防止了幻读又提升了并发。

> 虽然这个锁的名字叫插入意向锁，但是它是行级锁，跟意向锁这个表级锁，是没什么关系的。

### AUTO-INC 锁

AUTO-INC锁是一种特殊的表级锁，由插入带有AUTOINCREMENT 列的表的事务获取。 在最简单的情况下.如果一个事务正在向表中插入值，任何其他事务都必须等待，以便执行它们自己的插入操作，这样第一个事务插入的行就会接收到连续的主键值。

最主要是**用于保护AUTO\_INCREMENT列的自增值生成过程，确保并发插入时自增值的唯一性和连续性，防止多个事务生成相同的自增值**

分三种锁模式： 通过参数 `<span leaf="">innodb_autoinc_lock_mode</span>` 控制：


| 模式         | 值 | 描述           | 特点                 |
| ------------ | -- | -------------- | -------------------- |
| **传统模式** | 0  | 表级锁         | 安全性最高，性能最低 |
| **连续模式** | 1  | 轻量级互斥锁   | 平衡安全性和性能     |
| **无锁模式** | 2  | 预先分配自增值 | 性能最高，可能不连续 |

### 悲观锁

数据库中的锁，按照使用方式来分，也可分为**悲观锁**和**乐观锁**。

悲观锁的核心思想：\*\*"悲观"地认为数据一定会被其他事务修改，所以提前加锁**。所以在操作时是先获取锁，再进行业务操作。口诀就是“**一锁二查三更新\*\*”。

#### 悲观锁示例（行级锁）

```
-- 悲观锁示例
START TRANSACTION;

SELECT * FROM products WHERE id = 1 FOR UPDATE;  -- 加排他锁
UPDATE products SET stock = stock - 1 WHERE id = 1;

COMMIT;
```

在对id=1的记录修改前，先通过for update的方式进行加锁，然后再进行修改。这就是比较典型的悲观锁策略。

如果以上修改库存的代码发生并发，同一时间只有一个线程可以开启事务并获得id=1的锁，其它的事务必须等本次事务提交之后才能执行。这样我们可以保证当前的数据不会被其它事务修改。

#### 悲观锁示例（表级锁）

```
-- 表级锁
LOCK TABLES products WRITE;
-- 执行业务操作
......
-- 表解锁
UNLOCK TABLES;
```

当对products表进行操作的时候，会先对表加上锁，然后再操作。这样其他业务就不可以在对表中的数据进行修改了。体现了悲观的思想，业务使用方感觉“总有刁民想谋害朕”，加了锁之后安全感暴增！

#### 悲观锁实际场景示例-金融交易（转账场景）

```
-- 转账操作：从账户A转100元到账户B
STARTTRANSACTION;

-- 锁定付款方账户
SELECT balance FROM accounts WHEREid = 1FORUPDATE;
-- 业务需先检查balance的余额是否大于等于100，然后再执行扣减
UPDATE accounts SET balance = balance - 100WHEREid = 1;

-- 锁定收款方账户  
SELECT balance FROM accounts WHEREid = 2FORUPDATE;
UPDATE accounts SET balance = balance + 100WHEREid = 2;

COMMIT;
```

#### 悲观锁实际场景示例-库存扣减（防止超卖）

```
-- 商品库存扣减
STARTTRANSACTION;

SELECT stock FROM products WHEREid = 1FORUPDATE;

-- 检查库存是否充足
-- 如果充足则扣减
UPDATE products SET stock = stock - 1WHEREid = 1;

COMMIT;
```

### 乐观锁

**乐观锁总是"乐观"地认为数据不会被其他事务修改，不加锁直接操作。 会先进行操作，提交时检查是否发生冲突。主要通过版本号、时间戳等方式检测冲突。**

> MySQL 本身不提供内置乐观锁，需要用户在表结构设计或应用层通过 CAS 思想自行实现。CAS本质就是一项乐观锁技术，当多个线程尝试使用CAS同时更新同一个变量时，只有其中一个线程能更新变量的值，而其它线程都失败，失败的线程并不会被挂起，而是被告知这次竞争中失败，并可以再次尝试。

#### 悲观锁示例

```
-- 1. 查询数据时获取版本号
SELECTid, name, stock, versionFROM products WHEREid = 1;

-- 2. 更新时检查版本号
UPDATE products 
SET stock = stock - 1, version = version + 1
WHEREid = 1ANDversion = 旧版本号;

-- 3. 检查更新是否成功
-- 如果 update 成功记录数为0，说明数据被其他事务修改，版本号已经不再试第一步查出来的那个版本号了。
```

完整的示例

```
-- 创建带有版本号的表
CREATETABLE products (
    idINT PRIMARY KEY,
    nameVARCHAR(50),
    stock INT,
    versionINTDEFAULT1
);

-- 乐观锁更新过程
STARTTRANSACTION;

-- 步骤1：读取数据和版本号
SELECTid, stock, versionFROM products WHEREid = 1;
-- 假设获取到：version = 5, stock = 100

-- 步骤2：执行业务逻辑（在应用程序中）
-- 计算新库存：new_stock = 100 - 1 = 99

-- 步骤3：更新数据并检查版本号
UPDATE products 
SET stock = 99, version = 6WHEREid = 1ANDversion = 5;

-- 步骤4：检查更新结果
-- 如果 update 成功记录数为1，更新成功
-- 如果 update 成功记录数为0，数据被其他事务修改，需要重试

COMMIT;
```

在实际使用过程中，可以直接用更新时间的时间戳来实现乐观锁。示例如下：

```
-- 使用时间戳实现乐观锁
-- 查询出更新时间
SELECT id, stock, updated_at FROM products WHERE id = 1;

-- 用更新时间戳做版本号，再将现在时间更新为最新版本
UPDATE products 
SET stock = stock - 1, updated_at = NOW() 
WHERE id = 1 AND updated_at = '2025-01-16 10:30:00';
```

在实际业务中重试都是在业务层处理的，示例如下：

```
// Java中的乐观锁重试机制
public boolean updateProductStock(Long productId, int quantity) {
    int maxRetries = 3;
    
    for (int i = 0; i < maxRetries; i++) {
        // 1. 查询数据和版本号
        Product product = productMapper.selectById(productId);
        
        // 2. 尝试更新
        int affectedRows = productMapper.updateStock(
            productId, 
            product.getStock() - quantity,
            product.getVersion()
        );
        
        // 3. 检查更新结果
        if (affectedRows > 0) {
            returntrue; // 更新成功
        }
        
        // 4. 更新失败，短暂休眠后重试
        Thread.sleep(100);
    }
    
    returnfalse; // 重试多次后仍然失败
}
```

### 悲观锁与乐观锁总结

#### 悲观锁与乐观锁对比

![](https://mmbiz.qpic.cn/mmbiz_png/51zXejl2tKzSiar6k49nriaHRIfGyNoGKicD3ibuoDNDDCnnDRYibeHzEgiabvDNnvqq7AnA9iaG8hso6o2lFVT83qXRA/640?wx_fmt=png&from=appmsg&watermark=1)

#### 悲观锁与乐观锁实际使用中的选型建议

##### 悲观锁适用场景

* 写操作频繁（写占比 > 30%）
* 冲突概率高的业务
* 临界区执行时间长（包含IO操作）
* 强一致性要求（金融交易）

##### 乐观锁适用场景

* 读操作频繁（读占比 > 80%）
* 冲突概率低的业务
* 临界区执行时间短（纯内存操作）
* 系统吞吐量要求高
* 可以接受偶尔的重试

##### 应用案例

案例一：电商商品库存

```
-- 悲观锁方案（适合秒杀场景）
STARTTRANSACTION;
SELECT stock FROM products WHEREid = 1FORUPDATE;

IF (stock > 0) THEN
    UPDATE products SET stock = stock - 1WHEREid = 1;
    -- 创建订单...
ENDIF;
COMMIT;

-- 乐观锁方案（适合普通商品）
UPDATE products 
SET stock = stock - 1, version = version + 1
WHEREid = 1ANDversion = #{version} AND stock > 0;

IF (affected_rows = 1) THEN
    -- 创建订单...
ELSE
    -- 库存不足或版本冲突，提示用户
ENDIF;
```

案例二：用户账户余额

```
-- 悲观锁方案（转账场景）
STARTTRANSACTION;
SELECT balance FROM accounts WHEREid = 1FORUPDATE;

IF (balance >= amount) THEN
    UPDATE accounts SET balance = balance - amount WHEREid = 1;
    -- 记录流水...
ENDIF;
COMMIT;

-- 乐观锁方案（积分兑换）
UPDATE accounts 
SET points = points - cost, version = version + 1
WHEREid = 1
ANDversion = #{version} 
AND points >= cost;
```

##### 最佳方案

实际使用中可以监控业务的冲突概率来作为依据从而动态调整锁策略。 用Java实现一个使用例子（不含实现）：

```
// 智能锁选择策略
public void updateData(Data data) {
    // 根据冲突概率动态选择锁策略
    if (getConflictRate() > 0.1) {  // 冲突率 > 10%
        // 切换到悲观锁
        usePessimisticLock(data);
    } else {
        // 使用乐观锁
        useOptimisticLock(data);
    }
}
```

在这种方案中不光要监控冲突率，还有其他需要处理的细节点。

* **监控冲突率：根据实际情况动态调整锁策略**
* **设置重试上限：乐观锁避免无限重试**
* **降级机制：乐观锁失败次数过多时降级到悲观锁**
* **版本号管理：确保版本号字段的索引优化**
* **死锁预防：悲观锁场景下按固定顺序加锁**

### 总结

MySQL InnoDB 的锁体系可以概括为“**两类思想（悲观 vs 乐观）**、**三种粒度（表、页、行）**、**四种常用算法（Record / Gap / Next-Key / Insert Intention）**”。 理解它们的核心是抓住两条主线：

1. **隔离级别决定锁的“范围”——RR 用 Next-Key Lock 防幻读，RC 用 Record Lock 提并发；**
2. **索引决定锁的“位置”——锁永远加在索引上，无索引则退化为表锁或隐藏聚簇索引锁。**

实际选型时，先量化冲突概率与一致性要求：

* 冲突高、临界区长、强一致 → 悲观锁（FOR UPDATE / LOCK TABLES）；
* 冲突低、读多写少、可重试 → 乐观锁（CAS+版本号/时间戳）。

最后别忘了监控：把“冲突率、死锁次数、重试成功率”做成仪表盘，让锁策略随业务流量动态伸缩，才是真正的“高并发正确姿势”。
