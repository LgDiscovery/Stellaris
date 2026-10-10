# 04 · Spring 全家桶（15 题）

> **模块主线**：Spring 全家桶的本质是「**用代理 + 容器把横切关注点从业务代码里抽走**」——IoC/AOP 是底座，SpringBoot 解决「怎么少写配置」，SpringCloud 解决「多个 Boot 应用怎么协同」，MyBatis 解决「ORM 怎么被容器管起来」。所有面试题最终都落在三个问题上：**Bean 什么时候被创建、方法调用有没有被代理拦住、跨进程/跨线程时上下文还在不在**。

## 模块导览

> **背景**：Spring 是 Java 后端唯一无法绕开的框架，面试官问它不是为了考 API，而是通过它判断三件事——你是否理解框架的**运行时机制**（能定位线上问题）、是否踩过**代理失效**这类高频坑（够不够实战）、能否把 Spring 当成**架构积木**去解耦（有没有架构视角）。9 年经验的候选人如果只会背「Bean 生命周期八股」，答不出三级缓存为什么是三级、@Transactional 什么时候失效，反而会显得经验「注水」。
>
> 尤其要注意：**@Transactional 失效**和**自动装配原理**是高级岗的分水岭题，前者考实战踩坑，后者考框架源码。期权风控这类监管级系统对「配置不生效、事务不回滚」是零容忍的，所以面试官一定会往这条线上深挖。

**学完这 15 题，你能获得什么：**
- 能把 Bean 生命周期、三级缓存、AOP 代理选择规则串成一条线讲清楚，而不是背八股；
- 能准确列出 @Transactional 失效的 8 种场景并给出生产级修复方案（这是高级岗核心分水岭）；
- 能从源码层面讲清 SpringBoot 自动装配（含 2.7→3.x SPI 文件演进）并手写一个 Starter；
- 能讲清 SpringCloud 五大组件（Gateway/OpenFeign/Sentinel/Apollo/链路追踪）在微服务里各解决什么问题；
- 能把 MyBatis 执行链路、缓存失效、Mapper 动态代理讲透，并回答 MyBatis-Plus 的取舍。

**覆盖技术点速查：**

| 分层 | 技术点 |
|------|--------|
| 基础 | IoC/DI 三种注入、Bean 生命周期、BeanFactory vs FactoryBean vs ApplicationContext、AOP 原理与通知顺序、SpringMVC 请求流程 |
| 进阶 | 三级缓存与循环依赖、@Transactional 失效 8 场景、事务传播 7 种、自动装配 SPI 演进、启动流程与优化、MyBatis 四大组件与缓存 |
| 高阶 | OpenFeign + Sentinel 治理、Apollo 推拉结合与灰度、Gateway 基于 WebFlux 的取舍、TraceId 透传、SchedulerX/XXL-Job 调度原理、微服务拆分落地 |

**复习指引**：一面（基础）重点考 Q1–Q5、Q6、Q8；二面（实战）重点考 Q6、Q7、Q9、Q10、Q11；三面（架构）重点考 Q12–Q15。关联模块：Q7 与「**MySQL 事务/锁**」模块强关联（Spring 事务只是把 MySQL 事务的开启/提交/回滚包了一层 AOP）；Q11 与「**消息队列 / CAP**」模块关联（熔断降级是可用性取舍）；Q12 与「**分布式组件选型**」模块关联（配置中心是选型的一环，本模块只讲 Apollo 的使用与原理）。

**本模块目录（15 题）：**
1. [题 1：Bean 的生命周期，@PostConstruct 和 afterPropertiesSet 谁先执行？](#q1)
2. [题 2：依赖注入三种方式 + 循环依赖，三级缓存为什么是三级？](#q2)
3. [题 3：BeanFactory、FactoryBean、ApplicationContext 有什么区别？](#q3)
4. [题 4：AOP 原理，JDK 代理和 CGLIB 怎么选？拦截器、过滤器、AOP 有什么区别？](#q4)
5. [题 5：一次 HTTP 请求在 SpringMVC 里是怎么走完的？@RequestBody 什么原理？](#q5)
6. [题 6：@Transactional 什么时候会失效？（高频重点）](#q6)
7. [题 7：事务的 7 种传播行为，Spring 事务为什么能回滚？](#q7)
8. [题 8：SpringBoot 自动装配原理，你自己写过 Starter 吗？](#q8)
9. [题 9：SpringBoot 的启动流程是怎样的？线上怎么优化启动速度？](#q9)
10. [题 10：MyBatis 的执行流程，`#{}` 和 `${}` 有什么区别？](#q10)
11. [题 11：微服务拆分后 OpenFeign 调用怎么设置超时和熔断降级？](#q11)
12. [题 12：Apollo 配置中心原理，风控规则改完怎么做到秒级甚至毫秒级生效？](#q12)
13. [题 13：Gateway 断言和过滤器怎么工作？TraceId 是怎么跨服务透传的？](#q13)
14. [题 14：分布式任务调度（XXL-Job/SchedulerX）原理，多环境主备怎么切换？](#q14)
15. [题 15：期权风控实时预警系统，你用 Spring 全家桶怎么落地？](#q15)

**本模块图表索引：**

- **图 4-1**：Bean 的完整生命周期 —— 题 1
- **图 4-2**：三级缓存解决循环依赖的时序 —— 题 2
- **图 4-3**：为什么必须用三级缓存而非二级缓存 —— 题 2
- **图 4-4**：AOP 代理的选择规则 —— 题 4
- **图 4-5**：过滤器、拦截器、AOP 的执行顺序 —— 题 4
- **图 4-6**：SpringMVC 一次请求的完整流程 —— 题 5
- **图 4-7**：@Transactional 失效的判定决策树 —— 题 6
- **图 4-8**：事务传播行为 7 种语义对比 —— 题 7
- **图 4-9**：Spring 事务回滚的实现链路 —— 题 7
- **图 4-10**：SpringBoot 自动装配的完整链路 —— 题 8
- **图 4-11**：自定义 Starter 的结构 —— 题 8
- **图 4-12**：SpringApplication.run 的启动流程 —— 题 9
- **图 4-13**：MyBatis 的执行流程 —— 题 10
- **图 4-14**：微服务调用链与熔断降级 —— 题 11
- **图 4-15**：Apollo 配置热更新的推拉结合机制 —— 题 12
- **图 4-16**：期权风控系统的微服务架构总览 —— 题 15

---

## 基础考察题（5 道）

### <a id="q1"></a>题 1：Bean 的生命周期，@PostConstruct 和 afterPropertiesSet 谁先执行？

- **题干**：「你简历写了熟练 Spring，那先说说 Bean 的生命周期吧，创建到销毁都经历了什么？我再问一句，@PostConstruct 和 InitializingBean 的 afterPropertiesSet 哪个先跑？」
- **考察点**：对 IoC 容器运行时的理解深度 + 有没有真的看过源码（BeanPostProcessor 的插入点）+ 能不能把「AOP 代理在哪里生成」跟生命周期挂钩。

**标准答案**：

1. **实例化（Instantiation）**：容器通过构造器或工厂方法（`@Bean` / `FactoryBean.getObject()`）创建 Bean 的**原始对象**，此时对象里所有属性都是 null。对应源码 `AbstractAutowireCapableBeanFactory.createBeanInstance()`。
2. **属性填充（Populate）**：`populateBean()` 做依赖注入，把依赖的 Bean 引用注入进来。注意——**AOP 代理在这一步还没生成**，属性填充的是别的 Bean。
3. **Aware 接口回调**：依次回调 `BeanNameAware`、`BeanClassLoaderAware`、`BeanFactoryAware`；如果是 ApplicationContext 容器，还会通过 `ApplicationContextAwareProcessor` 回调 `EnvironmentAware`、`ApplicationContextAware`、`BeanNameAware` 等。
4. **BeanPostProcessor.postProcessBeforeInitialization**：所有 `BeanPostProcessor` 的前置处理。**`@PostConstruct` 就是在这一步被 `CommonAnnotationBeanPostProcessor` 执行的**。
5. **初始化（Initialization）**：顺序是 `@PostConstruct` → `InitializingBean.afterPropertiesSet()` → 自定义 `init-method`（`@Bean(initMethod="...")` 或 `@PostConstruct` 同级）。所以**`@PostConstruct` 一定早于 `afterPropertiesSet`**。
6. **BeanPostProcessor.postProcessAfterInitialization**：**AOP 代理在这里生成**（`AbstractAutoProxyCreator.wrapIfNecessary()` 判断是否需要增强，需要就返回代理对象，替换掉原始对象）。
7. **就绪使用**：Bean 进入一级缓存 `singletonObjects`，可以被注入、被调用。
8. **销毁**：容器关闭时（`ctx.close()` / `DisposableBeanAdapter.destroy()`），顺序是 `@PreDestroy` → `DisposableBean.destroy()` → 自定义 `destroy-method`。注意：**prototype 作用域的 Bean，Spring 不管理销毁**。

一句话记忆：**实例化 → 属性填充 → Aware → BPP.before → @PostConstruct/afterPropertiesSet/init-method → BPP.after(生成代理) → 使用 → @PreDestroy/destroy/destroy-method**。

**图 4-1：Bean 的完整生命周期**

```mermaid
flowchart TD
    N1["1 实例化<br/>构造器或工厂方法 new 出原始对象"] --> N2["2 属性填充 populateBean<br/>依赖注入 此时还没有 AOP 代理"]
    N2 --> N3["3 Aware 接口回调<br/>BeanNameAware BeanClassLoaderAware BeanFactoryAware"]
    N3 --> N4["4 BeanPostProcessor<br/>postProcessBeforeInitialization"]
    N4 --> N5["5 初始化第一步<br/>@PostConstruct"]
    N5 --> N6["6 初始化第二步<br/>InitializingBean.afterPropertiesSet"]
    N6 --> N7["7 初始化第三步<br/>自定义 init-method"]
    N7 --> N8["8 BeanPostProcessor<br/>postProcessAfterInitialization<br/>AOP 代理在此生成"]
    N8 --> N9["9 就绪使用<br/>进入一级缓存 singletonObjects"]
    N9 --> N10["10 销毁第一步<br/>@PreDestroy"]
    N10 --> N11["11 销毁第二步<br/>DisposableBean.destroy"]
    N11 --> N12["12 销毁第三步<br/>自定义 destroy-method"]
```

**企业实际处理**：
- Spring Boot 3.x / JDK 17 环境，业务代码基本不用 `InitializingBean`，统一用 `@PostConstruct`（或构造器 + `@Autowired` 做初始化），因为 `@PostConstruct` 符合 JSR-250 标准、可读性好。
- 需要「Bean 创建完成后做一件事」的场景，团队约定用 `ApplicationRunner` / `CommandLineRunner`（容器 refresh 完成后执行）或 `ApplicationListener<ApplicationReadyEvent>`，而不是往 `@PostConstruct` 里塞重逻辑——因为 `@PostConstruct` 阶段容器还没完全就绪，此时发起远程调用、启动定时任务容易出问题。
- 上线前用 Actuator 的 `/actuator/beans`、`/actuator/conditions` 排查 Bean 是否被创建、哪些自动配置生效。

**为什么这样做**：
- **因为 Aware / BPP / init 三段拆开是为了「扩展点分层」**：Aware 给 Bean 拿到容器基础设施；BPP 给框架（AOP、`@PostConstruct`、`@Async` 校验）插手；init 才是用户自己的初始化逻辑。分层是为了让「框架扩展」和「业务初始化」互不污染。
- **因为 AOP 代理必须放在 BPP.after 而不是实例化后立刻生成**：如果实例化就代理，那么属性填充、Aware 回调、init-method 全都会作用在代理对象上，而代理对象的方法调用会走切面，导致 Bean 自身初始化逻辑被自己的切面拦截，语义就乱了。放在 after 阶段，保证「Bean 自己先完整地活一遍，最后再穿上代理这件外套」。
- **因为 prototype 不管理销毁是权衡**：容器不持有 prototype 引用（否则内存泄漏），既然不持有也就无法回调销毁方法，所以只能交给调用方或 GC。这是「内存安全」优先于「生命周期完整性」的取舍。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：演示生命周期回调顺序与「别在 @PostConstruct 里做重活」的正确替代

```java
@Component
public class RiskEngineWarmUpBean implements InitializingBean, DisposableBean {

    private static final Logger log = LoggerFactory.getLogger(RiskEngineWarmUpBean.class);
    private final RuleRepository ruleRepository;

    public RiskEngineWarmUpBean(RuleRepository ruleRepository) {
        this.ruleRepository = ruleRepository;
        log.info("1. 构造器执行（实例化）");
    }

    @PostConstruct
    public void postConstruct() {
        log.info("2. @PostConstruct 先执行（仅做轻量本地校验）");
    }

    @Override
    public void afterPropertiesSet() {
        log.info("3. afterPropertiesSet 后执行");
    }

    @Override
    public void destroy() {
        log.info("5. destroy 执行（容器关闭）");
    }
}

/**
 * 重逻辑放到 ApplicationReadyEvent：此时容器已 refresh 完成，
 * 远程调用/线程池/数据源全部就绪，才是「预热风控规则」的正确时机。
 */
@Component
public class RiskRulePreloader implements ApplicationListener<ApplicationReadyEvent> {
    private final RuleRepository ruleRepository;

    public RiskRulePreloader(RuleRepository ruleRepository) {
        this.ruleRepository = ruleRepository;
    }

    @Override
    public void onApplicationEvent(ApplicationReadyEvent event) {
        // 预加载全量风控规则进本地 Caffeine 缓存，避免首笔预警请求抖动
        ruleRepository.loadAllIntoLocalCache();
    }
}
```

**面试加分项**：
- 主动补一句「`BeanPostProcessor` 本身是容器级组件，它只对**它注册之后**创建的 Bean 生效」——所以自定义 BPP 如果直接注册，可能对某些早创建的 Bean 失效。
- 提一句 `@PostConstruct` 底层是 `CommonAnnotationBeanPostProcessor`，`@Autowired` 底层是 `AutowiredAnnotationBeanPostProcessor`，两者都是 BPP，一个是 before 阶段、一个是属性填充阶段，压根不在同一环节。
- 如果目标 Bean 被 AOP 代理，`@PostConstruct` 仍会在原始对象上执行，但注入到别的 Bean 里的是代理——这解释了「为什么在 `@PostConstruct` 里调用自己的 `@Transactional` 方法依然失效」（self-invocation，见 Q6）。

---

### <a id="q2"></a>题 2：依赖注入三种方式 + 循环依赖，三级缓存为什么是三级？

- **题干**：「Spring 有几种依赖注入方式？你平时用哪种？再问你一个经典的——A 依赖 B、B 依赖 A 的循环依赖，Spring 是怎么解决的？它为什么需要三级缓存，二级不行吗？那构造器注入的循环依赖能解决吗？」
- **考察点**：DI 的工程取舍 + 三级缓存源码细节（能区分「提前暴露」和「提前代理」）+ 能否讲清构造器循环依赖无解的根因。**这题是判断候选人「背过」还是「懂了」的核心题。**

**标准答案**：

**第一部分：三种注入方式**

| 方式 | 写法 | 优点 | 缺点 |
|------|------|------|------|
| 构造器注入 | 构造方法 + `final` 字段 | **官方推荐**；依赖不可变；强依赖；启动即暴露缺失依赖；天然规避循环依赖（直接报错） | 参数多时构造器臃肿 |
| Setter 注入 | `@Autowired` 修饰 setter | 适合可选依赖、可重新配置 | 依赖可为 null，可能 NPE |
| 字段注入 | `@Autowired` 修饰字段 | 代码最短 | **不推荐**：无法 `final`、无法单元测试（不启容器就是 null）、隐藏依赖、易诱发循环依赖 |

Spring 官方明确推荐**构造器注入**：依赖用 `final` 保证不可变、单测无需容器、依赖过多时从构造器参数数量就能看出「这个类职责太重了」的坏味道。

**第二部分：三级缓存与循环依赖**

三级缓存都在 `DefaultSingletonBeanRegistry` 里：

| 级别 | 字段 | 存什么 |
|------|------|--------|
| 一级 | `singletonObjects` | 完整可用的单例 Bean |
| 二级 | `earlySingletonObjects` | 提前暴露的「半成品」Bean（可能是代理） |
| 三级 | `singletonFactories` | `ObjectFactory` 工厂（lambda，内含 `getEarlyBeanReference` 逻辑） |

循环依赖 A↔B 的流程（字段/Setter 注入）：
1. 创建 A：实例化 A 得到**原始对象**；**在属性填充之前**，把 A 的 `ObjectFactory` 放进三级缓存。
2. 填充 A 的属性，发现需要 B → 去创建 B。
3. 创建 B：实例化 B、放 B 的 `ObjectFactory` 到三级缓存；填充 B 的属性，发现需要 A。
4. B 找 A：从一级找不到 → 从二级找不到 → **从三级拿到 A 的 `ObjectFactory`，调用 `getObject()`**。这里就是关键：`getObject()` 内部走 `getEarlyBeanReference()`，如果 A 需要被 AOP 代理，**在这里就提前生成 A 的代理对象**；然后把这个对象放进**二级缓存**，并删除三级缓存。
5. B 注入到提前暴露的 A（或 A 的代理），B 完成初始化进入一级缓存。
6. A 继续注入 B，A 完成初始化，进入一级缓存。

**图 4-2：三级缓存解决循环依赖的时序**

```mermaid
sequenceDiagram
    participant CT as 容器 getSingleton
    participant A as Bean A
    participant B as Bean B
    participant CACHE as 三级缓存

    CT->>A: 触发创建 A
    A->>A: 实例化得到 A 的原始对象
    A->>CACHE: 属性填充之前 把 A 的 ObjectFactory 放入三级缓存
    A->>B: 填充属性发现依赖 B 触发创建 B
    B->>B: 实例化得到 B 的原始对象
    B->>CACHE: 把 B 的 ObjectFactory 放入三级缓存
    B->>CACHE: 填充属性发现依赖 A 逐级查找
    CACHE-->>B: 从三级缓存取出 A 的 ObjectFactory 并调用 getObject
    Note over CACHE,B: getObject 内部走 getEarlyBeanReference<br/>A 需要 AOP 时在此提前生成代理<br/>生成的对象放入二级缓存 同时删除三级缓存
    B->>B: B 注入 A 的早期引用 完成初始化 进入一级缓存
    A->>A: A 继续注入 B 完成初始化 进入一级缓存
```

**为什么是三级，二级不够吗？**
- 如果只有二级缓存（且二级缓存直接存「提前暴露对象」），那么在**每个 Bean 实例化后**都必须立刻判断「它要不要被 AOP 代理」，要就提前生成代理塞进二级缓存。问题有两个：
  1. **破坏了正常生命周期语义**：正常情况下 AOP 代理是在 `BeanPostProcessor.postProcessAfterInitialization`（初始化完成后）才生成的，提前生成意味着切面可能在属性填充/init 阶段就被触发。
  2. **所有 Bean 都被迫提前处理**，哪怕它根本没有循环依赖，性能与语义双重浪费。
- **三级缓存的本质是「延迟 + 按需」**：三级缓存存的是一个 `ObjectFactory` 工厂（lambda），只有在**真的发生循环依赖、有人来取它**的时候，才会调用 `getEarlyBeanReference()` 去生成（或复用）代理。没有循环依赖的绝大多数 Bean，这个工厂永远不会被调用，代理依然老老实实在 BPP.after 阶段生成。也就是说：**三级缓存用「一层工厂」换来了「不破坏生命周期语义 + 按需代理」**。

**为什么 AOP 代理要靠三级缓存（提前暴露的必须是代理而不是原始对象）？**
- 如果 B 注入的是 A 的**原始对象**而 A 最终被代理，那么 B 持有的是原始引用：B 调用 A 的方法时**不会走 A 的切面**（`@Transactional`、`@Async`、自定义切面全部失效）。为了正确性，B 必须注入到 A 的**代理**。所以 `getEarlyBeanReference()` 必须在这里就把代理建出来——这就是为什么三级缓存不能退化成「只存原始对象」。

**图 4-3：为什么必须用三级缓存而非二级缓存**

```mermaid
flowchart TD
    Q["核心矛盾 B 要注入 A 而 A 最终会被 AOP 代理"] --> D{"提前暴露给 B 的应该是什么"}
    D -->|只放原始对象| BAD["B 拿到的是原始引用<br/>B 调用 A 时不走 A 的切面<br/>事务与自定义切面全部失效"]
    D -->|只用二级缓存 实例化后立刻判代理| BAD2["每个 Bean 都要提前判断是否代理<br/>破坏了 AOP 在 BPP.after 生成的语义<br/>没有循环依赖的 Bean 也被迫提前处理"]
    D -->|三级缓存 存的是 ObjectFactory| GOOD["只在真有人来取时才调用 getEarlyBeanReference<br/>按需生成代理 不破坏生命周期语义"]
    GOOD --> OK["B 注入到的是 A 的代理 切面正常生效"]
```

**构造器注入的循环依赖为什么解决不了？**
- 循环依赖的发生点在**实例化阶段**（`createBeanInstance` 调构造器），而三级缓存提前暴露的前提是「**对象已经 new 出来了、只是还没填属性**」。构造器注入时，要调 A 的构造器就必须要 B 的实例，要 B 的实例就必须要 A 的实例——**第一步就卡死，没有任何半成品可以提前暴露**，Spring 只能抛 `BeanCurrentlyInCreationException`。
- 解决办法：① 改用 Setter/字段注入；② 依赖上打 `@Lazy`（注入一个代理，首次真正调用时才 `getBean`，相当于把「创建时机」推迟到运行期）；③ 用 `ObjectProvider` / `ApplicationContext.getBean()` 延迟获取；④ **最根本的：重新设计，循环依赖本身就是坏味道**。

**企业实际处理**：
- Spring Boot 2.6 起**默认禁止循环依赖**（`spring.main.allow-circular-references=false`），启动直接报错。团队不要靠 `allow-circular-references=true` 糊过去，而是重构拆解循环依赖。
- 业务代码统一**构造器注入**（配合 Lombok `@RequiredArgsConstructor` + `private final`），循环依赖会在启动期直接暴露，而不是运行期才出幺蛾子。
- 期权风控里 `RiskRuleService` ↔ `AlertNotifyService` 这种双向依赖，正确做法是抽一个 `AlertNotifier` 接口下沉，让两边都依赖抽象，而不是互相持有实现。

**为什么这样做**：
- **因为「启动期失败」优于「运行期失败」**：构造器注入让循环依赖在启动时炸掉，逼你在开发阶段就解决；字段注入把问题藏到运行期，在监管系统里这是不可接受的风险。
- **因为三级缓存是「语义正确性」优先于「实现简单」**：二级缓存实现更简单，但必须以破坏 Bean 生命周期为代价；Spring 选择了多一层工厂来保住「AOP 在 BPP.after 生成」这条铁律。这是典型的「用一点复杂度和微小的内存开销，换语义的纯粹性」。
- **因为 `@Lazy` 是「延迟初始化」而非「解决循环」**：它只是把循环依赖的触发点从启动期推迟到第一次调用，如果两个 Bean 真的在运行期互相调用且形成死锁式的构造，依然会出问题——所以它治标不治本。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：演示「构造器注入 + 接口下沉」如何从根上消除循环依赖

```java
// ❌ 反模式：RiskRuleService 和 AlertService 互相注入，Spring Boot 2.6+ 直接启动失败
// @Service public class RiskRuleService { @Autowired private AlertService alertService; }
// @Service public class AlertService  { @Autowired private RiskRuleService riskRuleService; }

// 正确姿势：下沉接口，单向依赖
public interface AlertNotifier {          // 告警能力抽象
    void notify(AlertEvent event);
}

@Service
public class RiskRuleService {
    private final AlertNotifier alertNotifier;   // 依赖抽象，不依赖实现类

    public RiskRuleService(AlertNotifier alertNotifier) {
        this.alertNotifier = alertNotifier;
    }

    public void evaluate(RiskContext ctx) {
        RiskResult result = doEvaluate(ctx);
        if (result.isTriggered()) {
            alertNotifier.notify(AlertEvent.of(result));
        }
    }

    private RiskResult doEvaluate(RiskContext ctx) { /* 规则引擎计算 */ return RiskResult.none(); }
}

@Service
@RequiredArgsConstructor
public class AlertService implements AlertNotifier {
    private final AlertChannelRepository channelRepository;  // 只依赖仓储，不再依赖 RiskRuleService

    @Override
    public void notify(AlertEvent event) {
        channelRepository.findByEventType(event.type())
                .forEach(channel -> channel.send(event));
    }
}
```

**面试加分项**：
- 能说出二级缓存里存的「可能是原始对象、也可能是代理」——**如果没有循环依赖，二级缓存几乎不被用到**（一级就够了），它存在的意义就是配合三级缓存的「提前暴露」中转。
- 能指出 `getEarlyBeanReference()` 是 `SmartInstantiationAwareBeanPostProcessor` 接口的方法，`AbstractAutoProxyCreator` 实现了它——这是把「三级缓存」和「AOP」两题串起来的加分点。
- 能补充：`@Async`、`@Transactional` 失效和循环依赖是**同源问题**——都是「注入的是原始对象而非代理」，理解一个就能解释另一个。
- 提到 `allowCircularReferences` 和 `allowRawInjectionDespiteWrapping` 这两个开关，说明你读过 `AbstractAutowireCapableBeanFactory`。

---

### <a id="q3"></a>题 3：BeanFactory、FactoryBean、ApplicationContext 有什么区别？

- **题干**：「Spring 里 BeanFactory 和 FactoryBean，名字就差一个词，你能说说区别吗？再对比一下 ApplicationContext 和 BeanFactory。」
- **考察点**：概念辨析 + 对容器分层（BeanFactory 是最底层容器，ApplicationContext 是对其的扩展）的理解 + 实际使用经验（有没有写过 FactoryBean）。

**标准答案**：

1. **BeanFactory = 最底层的 IoC 容器接口**。核心方法就一个 `getBean()`。它**默认懒加载**（用到才创建），只提供最基础的 DI 能力。子类 `DefaultListableBeanFactory` 是真正干活的实现。

2. **FactoryBean = 一个「能生产对象的 Bean」**。它是一个普通 Bean，但它实现后，容器对它特殊对待：
   - `getBean("&myFactoryBean")` 拿的是 **FactoryBean 本身**（前面加 `&`）；
   - `getBean("myFactoryBean")` 拿的是 `FactoryBean.getObject()` 返回的**产品对象**。
   - 典型用途：创建过程复杂、无法简单 new 的对象。Spring 自己大量使用——`SqlSessionFactoryBean`（生产 `SqlSessionFactory`）、`ProxyFactoryBean`（生产 AOP 代理）。
   - 三个方法：`getObject()` / `getObjectType()` / `isSingleton()`。

3. **ApplicationContext = BeanFactory 的超集（面向应用）**。它继承 `BeanFactory`，但额外提供：
   - **国际化**（`MessageSource`）；**事件发布**（`ApplicationEventPublisher` / `ApplicationListener`）；
   - **资源加载**（`ResourcePatternResolver`，`classpath*:` 批量加载）；
   - **自动注册 BeanPostProcessor / BeanFactoryPostProcessor**（BeanFactory 需要手动 `addBeanPostProcessor`）；
   - **默认预实例化**（`refresh()` 时创建所有非懒加载单例，启动即暴露错误）。
   - 常见实现：`AnnotationConfigApplicationContext`（注解）、`ClassPathXmlApplicationContext`（XML）、SpringBoot 的 `AnnotationConfigServletWebServerApplicationContext`。

一句话对比：**BeanFactory 是「发动机」，ApplicationContext 是「整车」；FactoryBean 是「车床」，它本身是个 Bean，但它产出的是另一个对象。**

**企业实际处理**：
- 业务开发基本只接触 `ApplicationContext`（即 `@Autowired` 注入的容器、SpringBoot 启动的那个）。
- `FactoryBean` 的使用场景：**封装复杂的第三方对象创建**。比如把 Kafka Producer、动态多数据源的 `DataSource`、加密后的 `SqlSessionFactory` 用 `FactoryBean` 包起来，让它们像普通 Bean 一样被注入。
- 需要拿到容器本身时用 `ApplicationContextAware` 或直接注入 `ApplicationContext`，但团队规范通常禁止业务代码直接 `getBean()`（服务定位器反模式，破坏依赖可见性）。

**为什么这样做**：
- **因为 FactoryBean 解决的是「创建逻辑与使用逻辑分离」**：调用方只想 `@Autowired` 拿到一个 `SqlSessionFactory`，不想知道它内部解析了多少 XML、装了多少插件。把这个脏活封进 `FactoryBean.getObject()`，调用方就能像用普通 Bean 一样用它——这是封装思想在容器层面的体现。
- **因为 ApplicationContext 是「应用级容器」而 BeanFactory 是「基础容器」**：分层是为了让轻量场景（如 MyBatis 内嵌的 `SqlSessionFactoryBean` 所在的小容器、单元测试）能用最小实现，而完整应用用重一点的 ApplicationContext。**分层 = 让不同场景各取所需，而不是一刀切**。
- **因为 ApplicationContext 默认预实例化是取舍**：牺牲了启动速度（要创建所有单例），换来了「启动即暴露配置错误」和「运行时无锁快速 getBean」。SpringBoot 的**懒加载优化**（Q9）正是把这个默认反过来。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：用 FactoryBean 封装一个需要复杂初始化的风控外部通道客户端

```java
/**
 * 演示 FactoryBean：把「读配置 → 建连接池 → 认证 → 健康检查」这一堆脏活
 * 封进 getObject()，业务侧只需 @Autowired AlertChannelClient 像用普通 Bean 一样用。
 */
public class AlertChannelClientFactoryBean implements FactoryBean<AlertChannelClient>, InitializingBean {

    private String endpoint;
    private String token;
    private AlertChannelClient client;

    // setter 省略，由 Spring 从配置注入
    public void setEndpoint(String endpoint) { this.endpoint = endpoint; }
    public void setToken(String token) { this.token = token; }

    @Override
    public void afterPropertiesSet() throws Exception {
        // 复杂初始化：建连接池 + 认证握手 + 健康检查
        this.client = AlertChannelClient.builder()
                .endpoint(endpoint)
                .token(token)
                .build();
        this.client.authenticate();
    }

    @Override
    public AlertChannelClient getObject() {
        return this.client;                 // getBean("...") 拿到的是这个
    }

    @Override
    public Class<?> getObjectType() {
        return AlertChannelClient.class;
    }

    @Override
    public boolean isSingleton() {
        return true;                        // getBean("&...") 拿到的才是本 FactoryBean
    }
}
```

**面试加分项**：
- 能说出 `&` 前缀这个细节，直接证明写过或用过 FactoryBean。
- 能举出 Spring 源码里的 FactoryBean 例子（`SqlSessionFactoryBean`、`ProxyFactoryBean`、`LocalContainerEntityManagerFactoryBean`），比空谈概念有说服力。
- 能补一句 `BeanFactoryPostProcessor` 与 `BeanPostProcessor` 的区别：前者在**所有 BeanDefinition 加载完、Bean 实例化前**执行（可改 BeanDefinition，如 `PropertySourcesPlaceholderConfigurer` 解析占位符）；后者在**Bean 实例化过程中**执行（如 `AutowiredAnnotationBeanPostProcessor`）。

---

### <a id="q4"></a>题 4：AOP 原理，JDK 代理和 CGLIB 怎么选？拦截器、过滤器、AOP 有什么区别？

- **题干**：「Spring 的 AOP 底层是怎么实现的？JDK 动态代理和 CGLIB 有什么区别，Spring 怎么选？顺便问一下，过滤器和拦截器、还有 AOP，这三个经常有人搞混，你能说清它们的区别和执行顺序吗？」
- **考察点**：AOP 底层机制 + 代理选择规则（含 SpringBoot 的默认行为）+ 对 Servlet/SpringMVC/AOP 三层拦截体系的理解。

**标准答案**：

**第一部分：两种代理**

| 维度 | JDK 动态代理 | CGLIB |
|------|--------------|-------|
| 原理 | 基于**接口**，`Proxy.newProxyInstance` 运行时生成实现同一接口的代理类 | 基于**继承**，用 ASM 字节码生成目标类的**子类**，覆写方法 |
| 前提 | 目标类**必须实现接口** | 目标类**不能是 final**；方法不能 `final`/`private`/`static` |
| 代理与目标关系 | 兄弟（同接口） | 父子（子类） |
| 性能 | 创建快、调用略慢（反射） | 创建慢（生成字节码）、调用快 |

**选择规则**：
- 传统 Spring（XML 时代）：目标类**有接口 → JDK 代理**；**无接口 → CGLIB**。
- **Spring Boot 2.x 起默认走 CGLIB**：`AopAutoConfiguration` 里 `spring.aop.proxy-target-class` 默认 `true`。原因：CGLIB 能代理到接口里没有的方法（如 `@Transactional` 标在实现类方法上），也避免了「注入类型不是接口时 JDK 代理 `ClassCastException`」的坑。
- 强制指定：`@EnableAspectJAutoProxy(proxyTargetClass = true/false)`。

**图 4-4：AOP 代理的选择规则**

```mermaid
flowchart TD
    S["目标 Bean 需要被 AOP 增强"] --> Q1{"是否显式设置 proxyTargetClass"}
    Q1 -->|true| C1["强制使用 CGLIB 子类代理"]
    Q1 -->|false| J1["强制使用 JDK 接口代理"]
    Q1 -->|未设置| Q2{"运行环境"}
    Q2 -->|传统 Spring XML 时代| Q3{"目标类是否实现接口"}
    Q2 -->|Spring Boot 2.x 起 AopAutoConfiguration 默认 proxy-target-class 为 true| C2["默认使用 CGLIB"]
    Q3 -->|有接口| J2["JDK 动态代理 基于接口"]
    Q3 -->|无接口| C3["CGLIB 子类代理"]
    C1 --> LIM["CGLIB 无法代理 final 类以及 final private static 方法"]
    C2 --> LIM
    C3 --> LIM
    J1 --> LIM2["JDK 代理只能拦到接口中声明的方法"]
```

**@Aspect 通知顺序**（正常返回时）：
```
@Around(proceed 之前) → @Before → 【目标方法】 → @AfterReturning → @After → @Around(proceed 之后)
```
- 异常时：`@Around` → `@Before` → 目标抛异常 → `@AfterThrowing` → `@After` → `@Around` 的异常部分。
- 多个切面之间用 `@Order`（值越小越靠外）控制；**外层切面先进入、后退出**（洋葱模型）。

**第二部分：过滤器 vs 拦截器 vs AOP**

| 维度 | Filter（过滤器） | Interceptor（拦截器） | AOP |
|------|------------------|-----------------------|-----|
| 规范/归属 | Servlet 规范（`jakarta.servlet.Filter`） | SpringMVC（`HandlerInterceptor`） | Spring AOP（`@Aspect`） |
| 作用位置 | **DispatcherServlet 之前/之后**，最外层 | HandlerMapping 找到 Handler 之后 | 任意被 Spring 管理的 Bean 方法 |
| 能否拿到 Handler 信息 | 不能（只有 `ServletRequest`） | 能（`preHandle` 的 `Object handler` 即 `HandlerMethod`） | 能（`JoinPoint` 拿到方法签名、参数） |
| 能否用 Spring Bean | 需 `FilterRegistrationBean` 注册才方便 | 能直接 `@Autowired` | 能 |
| 粒度 | URL | Handler（Controller 方法） | 方法 |
| 典型用途 | 编码、CORS、XSS、TraceId 入口埋点 | 登录鉴权、权限、日志 | 事务、缓存、限流、日志（业务方法级） |

**执行顺序**（一次请求）：
```
Filter#doFilter(before)
  → DispatcherServlet
    → Interceptor#preHandle
      → AOP @Around / @Before
        → Controller 方法
      → AOP @AfterReturning / @After
    → Interceptor#postHandle
  → Interceptor#afterCompletion
Filter#doFilter(after)
```
注意：`postHandle` 在视图渲染前执行，`afterCompletion` 在**整个请求完成后**（含渲染）执行，无论成功失败都会调。

**图 4-5：过滤器、拦截器、AOP 的执行顺序**

```mermaid
flowchart LR
    F1["Filter doFilter 前置逻辑<br/>最外层 拿到的是 ServletRequest"] --> I1["Interceptor preHandle<br/>拿得到 HandlerMethod"]
    I1 --> A1["AOP @Around 前半段 与 @Before"]
    A1 --> CTRL["Controller 目标方法"]
    CTRL --> A2["AOP @AfterReturning 与 @After"]
    A2 --> I2["Interceptor postHandle<br/>视图渲染前"]
    I2 --> I3["Interceptor afterCompletion<br/>请求完全结束 一定执行"]
    I3 --> F2["Filter doFilter 后置逻辑"]
```

**企业实际处理**：
- 期权风控实时预警服务里：**Filter** 用来做 TraceId 生成/提取和 CORS；**Interceptor** 做接口级鉴权和限流计数；**AOP** 做 `@Transactional`、接口耗时埋点、`@SentinelResource`。
- 团队规范：**不要用 AOP 做鉴权**（粒度用错、异常难统一），鉴权放 Interceptor 或网关；**不要用 Interceptor 做事务**（拿不到方法级切入点），事务用 AOP。
- 下单/预警这类**强事务**逻辑，AOP 代理必须生效——这就回到 Q2 的结论：不能自调用、不能 private、Bean 必须被容器管理。

**为什么这样做**：
- **因为三层拦截体系是「分层拦截、各司其职」**：Filter 在最外层，能做与 SpringMVC 无关的通用处理（编码、TraceId）；Interceptor 在 SpringMVC 内部，能拿到 Handler 元信息做鉴权；AOP 在方法级，能精细到参数与返回。**分层是为了让「通用能力放外层、业务能力放内层」，各自避免越界**。
- **因为 SpringBoot 默认 CGLIB 是「减少一类坑」的取舍**：CGLIB 代价是生成子类（轻微开销、不能代理 final），但换来「不强制目标类实现接口」和「代理到实现类自己的方法」。对以注解驱动为主的 SpringBoot 应用，这个取舍明显划算。
- **因为 AOP 基于继承的 CGLIB 决定了「private/final/static 方法切不到」**：子类无法覆写 private 方法，所以 `@Transactional` 标在 private 方法上必然失效——这直接是 Q6 的一个失效场景。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：一个统计风控规则执行耗时的切面，含多切面顺序控制

```java
@Aspect
@Component
@Order(1)                              // 数值越小越靠外层（先进入、后退出）
public class RiskRuleMetricsAspect {

    private static final Logger log = LoggerFactory.getLogger(RiskRuleMetricsAspect.class);

    // 只切风控规则引擎的 evaluate 方法
    @Around("@annotation(com.hongwang.risk.metrics.RuleMetric)")
    public Object measure(ProceedingJoinPoint pjp) throws Throwable {
        long start = System.nanoTime();
        String rule = pjp.getSignature().toShortString();
        try {
            Object result = pjp.proceed();          // 执行目标方法
            return result;
        } catch (Throwable t) {
            log.warn("风控规则执行异常: {}", rule, t);
            throw t;                                 // 必须原样抛出，否则事务/熔断感知不到
        } finally {
            long costMs = (System.nanoTime() - start) / 1_000_000;
            if (costMs > 200) {                      // 慢规则上报
                log.warn("慢规则 rule={} cost={}ms", rule, costMs);
            }
        }
    }
}

// 自定义注解，标在需要度量的方法上
@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
public @interface RuleMetric {
}
```

**面试加分项**：
- 能说出 Spring AOP **只支持方法级**，字段/构造器的「环绕」做不到（对比 AspectJ 的编译期织入、可到字段级）。
- 能提 `AopContext.currentProxy()` 暴露当前代理来解决自调用，但要注意需 `@EnableAspectJAutoProxy(exposeProxy = true)`，且**本质是绕开设计问题**，优先重构。
- 提到「同一个类里两个方法，外层无事务调内层有 `@Transactional` 的方法，内层失效」——直接连到 Q6。
- 知道 `@Order` 对 Filter 无效（Filter 顺序用 `@Order` 或 `FilterRegistrationBean.setOrder()`），Interceptor 顺序用 `InterceptorRegistry.order()`。

---

### <a id="q5"></a>题 5：一次 HTTP 请求在 SpringMVC 里是怎么走完的？@RequestBody 什么原理？

- **题干**：「说一说一个 HTTP 请求从进来到返回，在 SpringMVC 内部经历了哪些组件？@RequestBody 是怎么把 JSON 变成对象的？不带 @RequestBody 的参数又是怎么绑定的？」
- **考察点**：SpringMVC 核心组件流转 + 参数解析机制（HandlerMethodArgumentResolver）+ 消息转换器（HttpMessageConverter），能看出你是否理解「前端控制器模式」。

**标准答案**：

**完整流程（9 步）**：
1. **请求到达 `DispatcherServlet`**（前端控制器，SpringMVC 的总入口，在容器启动时注册）。
2. `DispatcherServlet` 通过 **`HandlerMapping`**（实际是 `RequestMappingHandlerMapping`）根据 URL + 请求方法 **匹配到 `HandlerExecutionChain`**——里面包含目标 `Handler`（即 `HandlerMethod`，封装了 Controller 实例和 Method）和一组 `HandlerInterceptor`。
3. **依次执行 `Interceptor#preHandle`**：任意一个返回 `false`，请求即中断（鉴权失败就走这里）。
4. `DispatcherServlet` 通过 **`HandlerAdapter`**（`RequestMappingHandlerAdapter`）适配并调用 Handler。这一步是核心：SpringMVC 用适配器模式把「各种类型的 Handler」统一成 `handle()` 调用。
5. **参数绑定**：`HandlerAdapter` 遍历方法参数，为每个参数找对应的 **`HandlerMethodArgumentResolver`**：
   - 普通 POJO / `@RequestParam` / `@PathVariable` → `ServletModelAttributeMethodProcessor` / `RequestParamMethodArgumentResolver`；
   - `@RequestBody` → `RequestResponseBodyMethodProcessor`；
   - `HttpServletRequest`、`Model` 等直接用 `ServletRequestMethodArgumentResolver`。
6. **执行 Controller 方法，得到返回值**。
7. **返回值处理**：通过 **`HandlerMethodReturnValueHandler`**：
   - 返回 `ModelAndView` → `ModelAndViewMethodReturnValueHandler`，交给 `ViewResolver` 解析视图、渲染；
   - `@ResponseBody` / `@RestController` → `RequestResponseBodyMethodProcessor`，用 **`HttpMessageConverter`** 把对象序列化成 JSON 写入响应体（`MappingJackson2HttpMessageConverter`）。
8. **执行 `Interceptor#postHandle`**（视图渲染前）。
9. **执行 `Interceptor#afterCompletion`**（请求完全结束后，无论成功失败），并处理异常（`HandlerExceptionResolver`）。

**图 4-6：SpringMVC 一次请求的完整流程**

```mermaid
flowchart TD
    REQ["HTTP 请求到达 DispatcherServlet"] --> HM["HandlerMapping 匹配到 HandlerExecutionChain<br/>含 HandlerMethod 与拦截器列表"]
    HM --> PRE["依次执行 Interceptor preHandle<br/>返回 false 则中断"]
    PRE --> HA["HandlerAdapter 适配并调用 Handler"]
    HA --> ARG["参数绑定 HandlerMethodArgumentResolver<br/>@RequestBody 走 RequestResponseBodyMethodProcessor"]
    ARG --> CTRL["执行 Controller 方法得到返回值"]
    CTRL --> RET["返回值处理 HandlerMethodReturnValueHandler<br/>@ResponseBody 用 HttpMessageConverter 序列化 JSON"]
    RET --> POST["执行 Interceptor postHandle<br/>视图渲染前"]
    POST --> DONE["执行 Interceptor afterCompletion<br/>请求完全结束后 一定执行"]
    DONE --> EX["异常由 HandlerExceptionResolver 处理"]
```

**@RequestBody 原理**：
- 参数解析器是 `RequestResponseBodyMethodProcessor`，它内部遍历 `HttpMessageConverter` 列表，找到能处理 `application/json` 的转换器（通常是 `MappingJackson2HttpMessageConverter`），调用 `read()` 用 Jackson 把请求体反序列化成对象。
- **前提：请求头必须 `Content-Type: application/json`**，否则会报 `HttpMediaTypeNotSupportedException`（415）——这是线上最常见的「@RequestBody 报 415」根因。
- 反向的 `@ResponseBody` 走同一个 `RequestResponseBodyMethodProcessor` 的 `write()`。

**不带 @RequestBody 的绑定原理**：
- 简单类型 + `@RequestParam` → 从 `request.getParameter()` 取；
- 复杂对象（无注解的 POJO）→ 走 `ModelAttribute` 机制，**从 query string / form-data 逐字段 setter 绑定**（`WebDataBinder` + `ConversionService` 做类型转换，如 String→Date）；
- `application/x-www-form-urlencoded` / `multipart/form-data` 是表单绑定，`application/json` 才是 `@RequestBody`——**两者不通用**，混用会拿到 null。

**企业实际处理**：
- 风控预警平台对外接口统一用 `@RestController` + `@RequestBody`（JSON），DTO 用 `@Valid` + `@NotNull` 做参数校验（`MethodArgumentNotValidException` 由 `@RestControllerAdvice` 统一转 400）。
- 统一响应包装 `Result<T>` + `@RestControllerAdvice` 全局异常（Q9），保证所有接口错误格式一致。
- 大报文（如批量规则导入）不走 `@RequestBody`（会占内存），改成流式读取或分片上传。

**为什么这样做**：
- **因为是「前端控制器模式 + 策略模式」的组合**：`DispatcherServlet` 是唯一入口（前端控制器），`HandlerMapping`/`HandlerAdapter`/`HandlerMethodArgumentResolver`/`HttpMessageConverter` 全是**策略接口**。这样新增一种参数解析方式、一种消息格式，只需加一个 Resolver/Converter，不用改核心流程——**开闭原则在框架层面的落地**。
- **因为参数解析与消息转换分离**：`ArgumentResolver` 负责「从请求里取原始数据」，`HttpMessageConverter` 负责「数据 ↔ 对象的序列化」。分离后，同一个 `@RequestBody` 能支持 JSON/XML/Protobuf（换 Converter 即可），**职责单一、可替换**。
- **因为 `afterCompletion` 才是资源释放的正确位置**：`postHandle` 在异常时不一定会执行（如果 Handler 抛异常，`postHandle` 被跳过），而 `afterCompletion` 一定会执行。所以 TraceId 清理、ThreadLocal 清理必须放 `afterCompletion`——放错地方就是 ThreadLocal 泄漏的经典 bug。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：一个能打印「到底走了哪个参数解析器/消息转换器」的拦截器 + 全局异常骨架

```java
@Component
public class RequestTraceInterceptor implements HandlerInterceptor {
    private static final Logger log = LoggerFactory.getLogger(RequestTraceInterceptor.class);

    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {
        if (handler instanceof HandlerMethod hm) {          // 拿得到 HandlerMethod 是拦截器相对过滤器的优势
            log.debug("进入 {}#{}, Content-Type={}",
                    hm.getBeanType().getSimpleName(), hm.getMethod().getName(),
                    request.getContentType());
        }
        return true;
    }

    @Override
    public void afterCompletion(HttpServletRequest request, HttpServletResponse response,
                                Object handler, Exception ex) {
        // 必须在这里清理，postHandle 在异常时会跳过，放那里会 ThreadLocal 泄漏
        TraceContext.clear();
    }
}

@RestController
@RequestMapping("/api/risk")
public class RiskRuleController {

    // @RequestBody：要求 Content-Type 为 application/json，否则 415
    @PostMapping("/rules")
    public Result<Void> createRule(@Valid @RequestBody RiskRuleDTO dto) {
        return Result.ok();
    }

    // 无 @RequestBody：走 @ModelAttribute / @RequestParam 绑定，来自 query string 或 form-data
    @GetMapping("/rules")
    public Result<List<RiskRuleDTO>> queryRules(@RequestParam String symbol,
                                                @RequestParam(defaultValue = "1") int page) {
        return Result.ok(List.of());
    }
}
```

**面试加分项**：
- 能说出「`HandlerMapping` 匹配不到会返回 404，匹配到但方法不允许会返回 405」——`DispatcherServlet` 对这两种情况有专门分支。
- 能提 `RequestMappingHandlerMapping` 在启动时会**扫描所有 `@RequestMapping` 方法并注册到 `MappingRegistry`**（含 URL 到 HandlerMethod 的映射），所以匹配是 O(1) 级别（`PathPattern` 匹配），不是每次遍历。
- 知道 Spring Boot 3 用 `PathPatternParser` 替代了旧的 `AntPathMatcher`，性能更好（预解析 pattern）。
- 提到 `HandlerExceptionResolver` 的注册顺序：`ExceptionHandlerExceptionResolver`（@ExceptionHandler）优先于 `ResponseStatusExceptionResolver`（@ResponseStatus）优先于 `DefaultHandlerExceptionResolver`（框架默认异常）。

---

## 进阶实战题（5 道）

### <a id="q6"></a>题 6：@Transactional 什么时候会失效？（高频重点）

- **题干**：「生产上你有没有遇到过加了 @Transactional 但事务没回滚的情况？说说 @Transactional 失效都有哪些场景，每种怎么排查、怎么改。这题你多讲点，我要看看你对事务的实战理解。」
- **考察点**：这是高级岗**最能拉开差距**的一题。面试官想确认：你是否真的在生产环境踩过事务坑，而不是只背过「事务有四大特性」。

**标准答案**（8 种失效场景，逐个讲根因 + 修复）：

1. **自调用（self-invocation）**：同一个类里方法 A 调方法 B，B 标了 `@Transactional`。
   - **根因**：事务靠 AOP 代理实现，`this.b()` 是**原始对象**内部调用，根本没经过代理，`TransactionInterceptor` 没机会介入。
   - **修复**：① 把 B 挪到另一个 Bean（首选）；② 注入自身代理 `@Autowired private XxxService self;` 然后 `self.b()`；③ `((XxxService) AopContext.currentProxy()).b()`（需 `exposeProxy=true`）；④ 用 `TransactionTemplate` 编程式事务。**注意：`@EnableAspectJAutoProxy(exposeProxy = true)` 才让 AopContext 可用。**
   - **特别注意：`@Transactional` 标在 Controller 上调用 Service 没问题，因为跨 Bean 走的是代理。**

2. **方法修饰符是 private**：`private void doXxx()` 标 `@Transactional`。
   - **根因**：CGLIB 靠生成子类覆写方法，**private 方法无法被覆写**；JDK 代理只在接口方法上生效。Spring 官方明确「`@Transactional` 只对 public 方法生效」。
   - **修复**：改成 public；或把逻辑抽到另一个 Bean 的 public 方法。

3. **方法非 public（protected / 包级默认）**：
   - **根因**：JDK 代理基于接口（接口方法天然 public），CGLIB 理论上能覆写 protected，但 **Spring 的 `AbstractFallbackTransactionAttributeSource.computeTransactionAttribute()` 明确跳过非 public 方法**，直接返回 null（不解析事务属性）。
   - **修复**：改 public。这也是为什么「内部工具方法想包事务」必须提到 public。

4. **异常被 catch 吞掉**：`try { ... } catch (Exception e) { log.error(...); }` 没有 rethrow。
   - **根因**：`TransactionInterceptor` 靠**捕获到异常**来决定 rollback，异常被吞就当成「正常返回」→ commit。
   - **修复**：要么 rethrow，要么在 catch 里手动 `TransactionAspectSupport.currentTransactionStatus().setRollbackOnly();`。

5. **异常类型不匹配**：抛的是**受检异常**（如 `IOException`、自定义 `BizException extends Exception`）。
   - **根因**：Spring 默认只对 **`RuntimeException` 和 `Error`** 回滚，受检异常默认 **commit**。
   - **修复**：`@Transactional(rollbackFor = Exception.class)`（团队规范应统一带上）。
   - 反例：自定义异常如果继承 `Exception` 而不是 `RuntimeException`，忘了写 `rollbackFor` 就一定不回滚——这是最隐蔽的坑之一。

6. **多线程 / 异步调用**：在事务方法里 `new Thread()` 或 `@Async` 里执行 DB 操作。
   - **根因**：事务上下文存在 **`TransactionSynchronizationManager` 的 ThreadLocal** 里（绑定的是 `ConnectionHolder`），子线程拿不到父线程的连接，**压根不在同一个事务**，回滚自然管不到子线程。
   - **修复**：让子线程的操作走独立事务（`REQUIRES_NEW`，但仍是独立连接、独立提交）；或用 `TransactionSynchronizationManager` 手动传递资源（不推荐，复杂）；根本方案是别在事务里开线程。

7. **传播行为配置不当**：例如外层方法**没有事务**，内层用 `NOT_SUPPORTED`；或期望内层独立回滚却用了默认 `REQUIRED`（内层异常被外层 catch，外层仍提交）；或外层 catch 了 `REQUIRES_NEW` 内层的异常导致外层误提交。
   - **根因**：对传播行为的语义理解错误。`REQUIRED`（默认）下内层和外层是**同一物理事务**，内层回滚会连累外层；要独立必须 `REQUIRES_NEW`。
   - **修复**：明确设计每个方法的传播语义；需要「内层失败不影响外层」用 `REQUIRES_NEW` 或 `NESTED`（savepoint）。

8. **类不被 Spring 管理**：`new XxxService().doSomething()`（自己 new 的）、类没被 `@Service`/`@Component` 扫描到、多数据源下没配对应的 `TransactionManager`。
   - **根因**：没有 Bean 就没有代理，没有代理就没有事务。
   - **修复**：确保被容器管理；多数据源时用 `@Transactional(transactionManager = "orderTxManager")` 指定事务管理器。

**额外常被追问的两条**（答出来是加分）：
- **表引擎不支持事务**：MySQL 的 **MyISAM 引擎不支持事务**（InnoDB 才支持），加了注解也是空转。
- **事务只读 + 大事务**：`@Transactional(readOnly = true)` 在某些驱动下只是 hint，不会真阻止写；另外大事务（长耗时）会导致锁持有时间长、连接占用，是**性能问题**而非「失效」，但常被一起问。

**图 4-7：@Transactional 失效的判定决策树**

```mermaid
flowchart TD
    S["@Transactional 方法没有回滚"] --> Q1{"Bean 是否被 Spring 管理且被代理"}
    Q1 -->|否 自己 new 或没被扫描到| F1["失效 场景 8 类不被容器管理"]
    Q1 -->|是| Q2{"方法是否 public"}
    Q2 -->|private 或非 public| F2["失效 场景 2 与 3 CGLIB 无法覆写 且 Spring 跳过非 public 方法"]
    Q2 -->|是| Q3{"调用是否经过代理"}
    Q3 -->|同类 this 自调用| F3["失效 场景 1 self-invocation 没走代理"]
    Q3 -->|经过代理| Q4{"异常是否被 catch 吞掉"}
    Q4 -->|是 没有 rethrow| F4["失效 场景 4 拦截器没感知异常 当成正常返回 commit"]
    Q4 -->|否 异常穿透| Q5{"异常类型是否匹配 rollbackFor"}
    Q5 -->|抛受检异常且未配 rollbackFor| F5["失效 场景 5 默认只回滚 RuntimeException 和 Error"]
    Q5 -->|匹配| Q6{"是否同一线程"}
    Q6 -->|新线程或 @Async| F6["失效 场景 6 ThreadLocal 不跨线程 不在同一事务"]
    Q6 -->|同一线程| Q7{"传播行为是否设计正确"}
    Q7 -->|默认 REQUIRED 却期望内层独立| F7["失效 场景 7 内外层是同一物理事务"]
    Q7 -->|正确| Q8{"数据库引擎是否支持事务"}
    Q8 -->|MyISAM| F8["失效 表引擎不支持事务 InnoDB 才支持"]
    Q8 -->|InnoDB 且全部通过| OK["事务正常回滚"]
```

**企业实际处理**：
- 团队编码规范强制：**所有 `@Transactional` 必须显式写 `rollbackFor = Exception.class`**（或用统一的自定义基类异常继承 RuntimeException），避免受检异常不回滚。
- **禁止在 `@Transactional` 方法内做 RPC / MQ / 长耗时操作**（如调告警通道、发短信），否则事务被拉长、连接池被打满。风控场景的正确姿势：**先落库（事务内）→ 事务外发告警**，或事务提交后再发（`TransactionSynchronizationManager.registerSynchronization` 的 `afterCommit`）。
- 关键操作（预警触发落库 + 状态变更）用**显式大事务 + 幂等键**，配合 `REQUIRES_NEW` 拆出「写操作日志」这种辅助动作，保证日志即使主事务回滚也留下。
- 上线用 Arthas 的 `trace` 或 Spring 的 `TransactionSynchronization` 埋点排查，定位「事务没生效」时第一件事就是看**调用方是不是同一个类**。

**为什么这样做**：
- **因为事务的本质是「AOP 拦截 + ThreadLocal 绑定连接」**：只要这条链上任一环节断了——代理没生效（1/2/3/8）、异常没传到拦截器（4/5/7）、线程换了（6）——事务就失效。把「事务失效」还原成「AOP 是否生效 + 异常是否穿透 + 线程是否同一」这三个问题，就不会漏场景。
- **因为默认只回滚 RuntimeException 是「设计取舍」**：受检异常在 Java 里表达「可恢复、调用方该处理的状况」（如文件不存在），Spring 认为这类异常调用方应自行处理而非一律回滚，所以默认不回滚。代价是**极易踩坑**，于是团队规范统一强制 `rollbackFor`。
- **因为不要把 RPC/MQ 放进事务是「资源占用」问题**：数据库连接是稀缺资源，事务持有连接直到提交；在事务内做 100ms 的远程调用，等于把连接借出去 100ms，高并发下连接池瞬间耗尽（风控实时系统 QPS 高，这点是致命的）。把「外部动作」移出事务边界是**典型的性能与稳定性优化**。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：正确的事务边界——写库在事务内、告警在提交后发，且强制 rollbackFor

```java
@Service
public class AlertTriggerService {

    private final AlertRecordRepository recordRepository;
    private final AlertNotifier alertNotifier;          // 跨 Bean，避免自调用

    public AlertTriggerService(AlertRecordRepository recordRepository, AlertNotifier alertNotifier) {
        this.recordRepository = recordRepository;
        this.alertNotifier = alertNotifier;
    }

    /**
     * 强制 rollbackFor：避免受检异常导致事务不回滚（失效场景 5）。
     * 只包住 DB 写操作，不包含远程调用（避免长事务占连接）。
     */
    @Transactional(rollbackFor = Exception.class)
    public void trigger(AlertEvent event) {
        int updated = recordRepository.markTriggered(event.ruleId(), event.symbol());
        if (updated == 0) {
            throw new BizException("预警记录状态已变更，幂等终止");   // RuntimeException，触发回滚
        }
    }

    /**
     * 对外入口：事务提交后再发告警。
     * 用 afterCommit 同步回调，保证「落库成功才告警」，且告警耗时不计入事务时长。
     */
    public void triggerAndNotify(AlertEvent event) {
        TransactionSynchronizationManager.registerSynchronization(new TransactionSynchronization() {
            @Override
            public void afterCommit() {
                alertNotifier.notify(event);            // 事务外的远程调用
            }
        });
        trigger(event);                                 // 跨 Bean 调用，代理生效
    }
}
```

**面试加分项**：
- 能主动说出「`@Transactional` 生效的**判断链路**：Bean 有没有被代理 → 方法是不是 public → 调用是否经过代理 → 异常是否穿透到拦截器 → 是否同一线程」，把 8 个场景归结成一条链，比零散罗列高一个层次。
- 提到 `TransactionTemplate` / `TransactionAspectSupport.currentTransactionStatus().setRollbackOnly()` 作为兜底手段，说明你实操过。
- 提到 `@Transactional` 默认 `propagation = REQUIRED`、`isolation = DEFAULT`（跟随数据库）、`timeout = -1`（不过期）、`readOnly = false`，说明你记得默认值。
- 提到「事务内 self-invocation 调用 `REQUIRES_NEW` 方法，注解不生效，因为压根没走代理」——把场景 1 和 7 串起来答，是高级岗水准。
- 结合业务：期权风控预警「触发落库 + 推送告警」，如果推送失败不能回滚落库（否则预警丢失），所以推送必须在事务外/afterCommit——**这个例子比抽象讲理论有说服力得多**。

---

### <a id="q7"></a>题 7：事务的 7 种传播行为，Spring 事务为什么能回滚？

- **题干**：「7 种事务传播行为你记得吗？各自什么语义？再往深问一层——Spring 事务底层到底是怎么回滚的？跟 MySQL 的事务是什么关系？」
- **考察点**：传播语义的准确记忆（尤其 REQUIRED vs REQUIRES_NEW vs NESTED）+ 能否把「Spring 事务」还原到「JDBC 连接 + ThreadLocal 绑定」这一层，看出你理解的是机制不是 API。

**标准答案**：

**7 种传播行为**（`Propagation` 枚举）：

| 传播行为 | 语义 | 何时用 |
|----------|------|--------|
| `REQUIRED`（默认） | 有事务则**加入**，无则**新建** | 绝大多数场景 |
| `REQUIRES_NEW` | **总是新建**独立事务，挂起当前事务 | 内层结果独立于外层（如写操作日志、记审计） |
| `SUPPORTS` | 有事务则加入，无则以**非事务**方式执行 | 查询方法，有没有事务都行 |
| `NOT_SUPPORTED` | 以**非事务**方式执行，挂起当前事务 | 不想受事务影响的耗时操作 |
| `MANDATORY` | 必须在事务中，**否则抛异常** | 强制「只能被事务方法调用」的内部方法 |
| `NEVER` | 必须**无事务**，否则抛异常 | 断言「此方法绝不能在事务里」 |
| `NESTED` | 有事务则**嵌套**（savepoint），无则等同 `REQUIRED` | 「内层失败只回滚内层，外层可继续」 |

**关键区分（高频）：`REQUIRES_NEW` vs `NESTED`**
- `REQUIRES_NEW`：开**全新物理事务**（新连接），挂起外层。内层提交/回滚**完全独立**，外层回滚**不影响已提交的内层**。
- `NESTED`：仍在外层事务内，只是打一个 **savepoint**。内层回滚**回退到 savepoint**，外层可继续；但**外层回滚会连内层一起回滚**（同一个物理事务）。
- 记忆：`REQUIRES_NEW` 是「另起炉灶、独立结算」，`NESTED` 是「同一笔账里打个标记、局部撤销」。

**图 4-8：事务传播行为 7 种语义对比**

```mermaid
flowchart TD
    ROOT["7 种传播行为 Propagation 枚举"] --> G1["REQUIRED 默认 有事务则加入 无则新建"]
    ROOT --> G2["REQUIRES_NEW 总是新建独立事务 挂起当前事务"]
    ROOT --> G3["NESTED 有事务则用 savepoint 嵌套 无则等同 REQUIRED"]
    ROOT --> G4["SUPPORTS 有则加入 无则以非事务方式执行"]
    ROOT --> G5["NOT_SUPPORTED 以非事务方式执行 挂起当前事务"]
    ROOT --> G6["MANDATORY 必须在事务中 否则抛异常"]
    ROOT --> G7["NEVER 必须无事务 否则抛异常"]
    G1 --> C1["外层回滚时 内层一起回滚"]
    G2 --> C2["外层回滚时 内层已提交 不受影响"]
    G3 --> C3["外层回滚时 内层一起回滚<br/>但内层回滚只回退到 savepoint 外层可继续"]
```

**Spring 事务为什么能回滚（底层机制）**：
1. **代理拦截**：`@Transactional` 被 `TransactionInterceptor`（实现 `MethodInterceptor`）拦截，方法调用先进入 `invoke()`。
2. **获取事务管理器**：`PlatformTransactionManager`（JDBC 场景是 `DataSourceTransactionManager`），`getTransaction()` 开启/加入事务。
3. **绑定连接（核心）**：`DataSourceTransactionManager.doBegin()` 从 `DataSource` 取一个 `Connection`，设 `autoCommit = false`，然后用 `TransactionSynchronizationManager.bindResource(dataSource, new ConnectionHolder(conn))` 把这个连接**绑定到当前线程的 ThreadLocal**。
4. **同一线程共享连接**：方法内所有 MyBatis/JdbcTemplate 的 DAO 操作，通过 `DataSourceUtils.getConnection(dataSource)` **从 ThreadLocal 里拿到同一个 Connection**——所以它们天然在**同一个事务**里。
5. **提交/回滚**：方法正常返回 → `commit()`；抛异常且命中回滚规则 → `rollback()`。最后 `cleanupAfterCompletion()` 解绑 ThreadLocal、归还连接（若打开）。
6. **回滚设置**：`TransactionAspectSupport.completeTransactionAfterThrowing()` 判断异常是否匹配 `rollbackFor`，匹配才 `rollback()`，否则 `commit()`。

**图 4-9：Spring 事务回滚的实现链路**

```mermaid
flowchart LR
    T["@Transactional 方法被调用"] --> P["AOP 代理 TransactionInterceptor 拦截 invoke"]
    P --> M["PlatformTransactionManager getTransaction"]
    M --> D["DataSourceTransactionManager doBegin"]
    D --> CONN["从 DataSource 取 Connection 设 autoCommit 为 false"]
    CONN --> BIND["bindResource 把 ConnectionHolder 绑定到当前线程 ThreadLocal"]
    BIND --> DAO["同线程 DAO 通过 DataSourceUtils 拿到同一 Connection"]
    DAO --> RES{"方法结果"}
    RES -->|正常返回| COMMIT["commit 提交"]
    RES -->|抛异常且匹配 rollbackFor| RB["rollback 回滚 底层靠 undo log"]
    COMMIT --> CL["cleanupAfterCompletion 解绑 ThreadLocal 归还连接"]
    RB --> CL
```

**与 MySQL 事务的关系**：
- Spring 事务**不是独立的事务实现**，它是**对 JDBC 事务（即数据库事务）的封装**。真正的提交/回滚由 MySQL 的 InnoDB 执行——`Connection.commit()` / `rollback()` 对应 MySQL 的 COMMIT/ROLLBACK。
- Spring 事务隔离级别默认 `DEFAULT` → **跟随数据库的隔离级别**（MySQL InnoDB 默认 `REPEATABLE READ`）。如果显式设置 `isolation`，Spring 会通过 `Connection.setTransactionIsolation()` 应用到该连接。
- **因此 MySQL 侧的一切规则照样适用**：锁（行锁/间隙锁）、MVCC 快照、undo log（回滚靠 undo log）、redo log（持久化）、两阶段提交（binlog + redo log）——这些是 MySQL 讲的，Spring 只是「发指令的人」。
- 一个直接推论：**`@Transactional` 的粒度 = 一个数据库连接的 commit 边界**。跨库（多数据源）默认不是原子事务，需要 XA / Seata 这类分布式事务方案。

**企业实际处理**：
- `REQUIRES_NEW` 的典型用途：**审计/操作日志**必须落库，即主业务回滚也要留痕 → 日志方法用 `REQUIRES_NEW`。但注意**它会多占一个数据库连接**（挂起外层、开新连接），高并发下要评估连接池大小。
- **禁止大事务**：风控规则批量导入几万条数据，如果一个大事务全包住，会长时间持锁 + 占连接。做法是**分批提交**（每 1000 条一个小事务），或异步化。
- 读方法统一 `@Transactional(readOnly = true)`：告诉驱动/框架这是只读，部分数据库能优化（MySQL 下主要是 hint，但对 ORM 的脏检查有优化，如 Hibernate 会跳过 flush）。
- 团队规范：**Controller 不开事务**，事务边界放在 Service 的 public 方法上，且**事务方法内不做 RPC**（呼应 Q6）。

**为什么这样做**：
- **因为「ThreadLocal 绑定连接」是 Spring 事务能跨多个 DAO 生效的根本**：如果没有这个绑定，每个 DAO 各自 `getConnection()` 会拿到不同连接，就不可能在同一事务里。ThreadLocal 让「同一线程的多次数据库操作」共享一个连接——这是**隐式传参**的精妙设计，代价是**跨线程就失效**（Q6 场景 6）。
- **因为 `REQUIRES_NEW` 要挂起外层是有成本的**：挂起意味着外层连接暂时不用但仍被占用，同时新开一个连接。这是「独立性」换「资源占用」的取舍，所以**不能滥用**——只有真正需要「独立提交/回滚」时才用。
- **因为 Spring 事务默认跟随数据库隔离级别是「不越权」的取舍**：Spring 不强行规定隔离级别，把选择权交给数据库（DBA 会按业务统一设定）。代价是「换数据库行为可能变」，所以对隔离级别敏感的业务必须显式声明。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：`REQUIRES_NEW` 记录审计日志 + `NESTED` 局部回滚

```java
@Service
public class RiskRuleImportService {

    private final RuleRepository ruleRepository;
    private final AuditLogService auditLogService;    // 跨 Bean，且内部 REQUIRES_NEW

    public RiskRuleImportService(RuleRepository ruleRepository, AuditLogService auditLogService) {
        this.ruleRepository = ruleRepository;
        this.auditLogService = auditLogService;
    }

    // 主事务：默认 REQUIRED
    @Transactional(rollbackFor = Exception.class)
    public void importRules(List<RiskRule> rules) {
        for (RiskRule rule : rules) {
            // 审计日志走 REQUIRES_NEW：即使主事务回滚，导入痕迹也保留（合规要求）
            auditLogService.log("IMPORT_RULE", rule.ruleId());
            ruleRepository.save(rule);
        }
    }
}

@Service
public class AuditLogService {

    private final AuditLogRepository auditLogRepository;

    public AuditLogService(AuditLogRepository auditLogRepository) {
        this.auditLogRepository = auditLogRepository;
    }

    // 独立事务：挂起调用方事务，自身提交/回滚互不影响
    @Transactional(propagation = Propagation.REQUIRES_NEW, rollbackFor = Exception.class)
    public void log(String action, String targetId) {
        auditLogRepository.insert(AuditLog.of(action, targetId, Instant.now()));
    }
}
```

**面试加分项**：
- 能说出「`MANDATORY` 和 `NEVER` 是用来**表达契约**的」——`MANDATORY` 表示「我只能被事务调用」（否则说明调用方漏了事务），`NEVER` 表示「我绝对不能进事务」（否则说明调用链设计错了），它们是**断言工具**，不是功能工具。
- 能补充 `NESTED` 依赖 `savepoint` 支持，**JDBC 3.0 才有的 `Connection.setSavepoint()`**，且**对某些数据库/驱动不生效**（Spring 会用 `JdbcTransactionObjectSupport` 判断是否支持）。
- 能说出「Spring 事务回滚后，`TransactionSynchronization` 的 `afterCompletion(STATUS_ROLLED_BACK)` 会被回调」——用来做补偿/清理。
- 把「Spring 事务」和「MySQL 三大日志（undo/redo/binlog）与两阶段提交」打通：**回滚靠 undo log，持久化靠 redo log**——面试官问「事务为什么能回滚」，从 Spring 答到 MySQL undo log 才算答完整。

---

### <a id="q8"></a>题 8：SpringBoot 自动装配原理，你自己写过 Starter 吗？

- **题干**：「SpringBoot 最核心的『约定优于配置』是怎么实现的？说说自动装配原理。@SpringBootApplication 这个注解里到底装了什么？还有——你写过自定义 Starter 吗？说说步骤。SpringBoot 3 在这方面有什么变化？」
- **考察点**：框架源码理解 + 对 2.7→3.x 关键变更的敏感度（新旧 SPI 文件路径）+ 有没有真正封装过 Starter（工程化能力）。

**标准答案**：

**第一步：`@SpringBootApplication` 是「三合一」**
```java
@SpringBootApplication
// = @SpringBootConfiguration（本质是 @Configuration，标识配置类）
// + @ComponentScan（扫描当前包及子包的 @Component，所以启动类要放根包）
// + @EnableAutoConfiguration（自动装配的开关）
```

**第二步：`@EnableAutoConfiguration` 也是「二合一」**
```java
@EnableAutoConfiguration
// = @AutoConfigurationPackage（把启动类所在包注册为「自动配置包」，供 JPA 实体扫描等使用）
// + @Import(AutoConfigurationImportSelector.class)   ← 自动装配的核心
```

**第三步：`AutoConfigurationImportSelector` 干了什么**
1. 它实现 `ImportSelector`，`selectImports()` 被 Spring 调用，返回**一堆自动配置类的全限定名**。
2. 通过 `getCandidateConfigurations()` → `ImportCandidates.load()` / `SpringFactoriesLoader.loadFactoryNames()` **读取 SPI 配置文件**，拿到所有候选自动配置类。
3. **去重 + 排除**：去掉重复的，去掉 `@SpringBootApplication(exclude = ...)` / `spring.autoconfigure.exclude` 指定的。
4. **过滤（关键）**：用 **`AutoConfigurationImportFilter`**（`OnClassCondition`、`OnBeanCondition`、`OnWebApplicationCondition`）做快速过滤——**只保留「条件满足」的配置类**。这一步在**不加载类**的前提下用 ASM 读注解元数据判断，非常快。
5. 最终把过滤后的配置类注册成 BeanDefinition，等待容器实例化。

**图 4-10：SpringBoot 自动装配的完整链路**

```mermaid
flowchart TD
    APP["@SpringBootApplication 三合一"] --> A1["@SpringBootConfiguration 本质是 @Configuration"]
    APP --> A2["@ComponentScan 扫描当前包及子包"]
    APP --> A3["@EnableAutoConfiguration 自动装配开关"]
    A3 --> B1["@AutoConfigurationPackage 注册自动配置包"]
    A3 --> B2["@Import AutoConfigurationImportSelector"]
    B2 --> C1["getCandidateConfigurations 读取 SPI 文件"]
    C1 --> C2["SpringBoot 2.6 及以前 读 META-INF/spring.factories"]
    C1 --> C3["SpringBoot 2.7 起 同时支持 imports 新文件<br/>旧 spring.factories 标记 deprecated"]
    C1 --> C4["SpringBoot 3.x 只认 org.springframework.boot.autoconfigure.AutoConfiguration.imports"]
    C2 --> D["去重 并排除 exclude 指定的配置"]
    C3 --> D
    C4 --> D
    D --> E["AutoConfigurationImportFilter 按条件注解快速过滤<br/>OnClassCondition OnBeanCondition OnWebApplicationCondition"]
    E --> F["注册过滤后的 BeanDefinition 等待容器实例化"]
```

**第四步：SPI 文件的演进（这题的核心变更点，必须答准）**

| 版本 | 文件路径 | 说明 |
|------|----------|------|
| SpringBoot ≤ 2.6 | `META-INF/spring.factories` | key = `org.springframework.boot.autoconfigure.EnableAutoConfiguration`，value = 逗号分隔的配置类 |
| SpringBoot 2.7 | **两者都支持** | 新增 `META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports`（一行一个类），`spring.factories` 中的自动配置标记为 **deprecated** |
| SpringBoot 3.x | **只认 imports 文件** | `spring.factories` 里的自动配置**不再生效**；自动配置类需用 **`@AutoConfiguration`** 注解标注（替代裸 `@Configuration`） |

- 新文件路径（务必记准）：`META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports`
- `spring.factories` 在 Boot 3 里仍存在，但**只用于 `ApplicationListener` / `ApplicationContextInitializer` / `EnvironmentPostProcessor` 等非自动配置场景**（即 `SpringFactoriesLoader` 的基础用法还在）。

**第五步：条件注解（自动装配的「开关」）**

`@Conditional` 家族（`@ConditionalOnXxx`，由 `OnClassCondition` / `OnBeanCondition` / `OnPropertyCondition` 等实现）：
- `@ConditionalOnClass`：classpath 上有某个类才生效（如没引 Redis 依赖，RedisAutoConfiguration 就不装配）；
- `@ConditionalOnMissingBean`：容器里**没有**某个 Bean 才装配（**这是让用户自定义 Bean 覆盖默认配置的关键**）；
- `@ConditionalOnProperty`：某配置项为指定值才生效（如 `spring.redis.enabled=true`）；
- `@ConditionalOnBean` / `@ConditionalOnWebApplication` / `@ConditionalOnResource` 等。

**第六步：自定义 Starter 的完整步骤**
1. 建两个模块：**`xxx-spring-boot-starter`**（空 jar，只做依赖聚合）+ **`xxx-spring-boot-autoconfigure`**（真正的自动配置代码）。
2. autoconfigure 模块里写 **`XxxProperties`**：`@ConfigurationProperties(prefix = "risk.alert")`，把可配置项收拢。
3. 写 **`XxxAutoConfiguration`**：用 **`@AutoConfiguration`**（SpringBoot 2.7+ / 3.x）或 `@Configuration`，配 `@EnableConfigurationProperties(XxxProperties.class)`，核心 Bean 上加 `@ConditionalOnMissingBean`（允许用户覆盖）、`@ConditionalOnClass`（依赖存在才装配）。
4. 在 `src/main/resources/META-INF/spring/` 下建文件 **`org.springframework.boot.autoconfigure.AutoConfiguration.imports`**，写入 `XxxAutoConfiguration` 的全限定名（一行一个）。
5. starter 模块的 `pom.xml` 依赖 autoconfigure 模块 + 需要的第三方依赖。
6. （可选）加 `META-INF/spring-configuration-metadata.json` 或 `spring-boot-configuration-processor` 让 IDE 有配置提示。

**图 4-11：自定义 Starter 的结构**

```mermaid
flowchart TD
    STARTER["xxx-spring-boot-starter<br/>空 jar 只做依赖聚合"] --> AUTO["xxx-spring-boot-autoconfigure<br/>真正的自动配置代码"]
    AUTO --> P["XxxProperties<br/>@ConfigurationProperties 收拢可配置项"]
    AUTO --> CFG["XxxAutoConfiguration<br/>@AutoConfiguration 配合条件注解"]
    CFG --> IMP["META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports<br/>一行一个自动配置类全限定名"]
    CFG --> BEAN["@Bean 配合 @ConditionalOnMissingBean 允许业务方覆盖默认实现"]
    CFG --> META["spring-configuration-metadata.json<br/>让 IDE 有配置提示 可选"]
```

**企业实际处理**：
- 公司内部统一封装：如「统一日志组件」「统一鉴权客户端」「风控规则 SDK」做成 Starter，各业务服务直接引依赖就能用，避免每个服务复制粘贴配置。
- Starter 里的 Bean 必须加 `@ConditionalOnMissingBean`，**否则业务方无法覆盖默认实现**——这是公司内部 Starter 最常见的设计缺陷。
- 排查「自动配置为什么没生效」：用 `--debug` 启动（打印 `ConditionEvaluationReport`）或 Actuator 的 **`/actuator/conditions`**（SpringBoot 3 里是 `/actuator/conditions`，旧版 `/actuator/autoconfig`），能看到「哪些自动配置生效/未生效，原因是 `@ConditionalOnClass` 没满足」。

**为什么这样做**：
- **因为 SPI + 条件注解 = 「自动装配」的两大支柱**：SPI 解决「框架怎么发现你的配置类」（解耦，不用硬编码 import）；条件注解解决「发现了但要不要生效」（按需装配）。**两者缺一不可**：只有 SPI 会加载一堆用不上的配置；只有条件注解没有 SPI，用户就得手动 `@Import`。
- **因为 `@ConditionalOnMissingBean` 是「约定优于配置但允许打破」的开关**：默认给一套能跑的配置，但用户一旦自己定义了 Bean，默认配置自动让位——这是「约定」与「自定义」的平衡点。**没有它，Starter 就是霸道不可覆盖的**。
- **因为 2.7→3.x 换 SPI 文件是「性能 + 明确性」的驱动**：`spring.factories` 是 Properties 格式，一个 key 下所有类挤在逗号分隔的值里，难读、难维护、加载要全量解析；`.imports` 一行一个类，**可读性和解析效率都更好**（`ImportCandidates` 用 `BufferedReader` 逐行读）。代价就是**不兼容老写法**，所以升级到 Boot 3 时所有自定义 Starter 都必须改文件。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：一个「风控告警通道」Starter 的自动配置骨架 + imports 文件内容

```java
// ===== autoconfigure 模块 =====

// 1. 配置属性类
@ConfigurationProperties(prefix = "risk.alert")
public class RiskAlertProperties {
    private boolean enabled = true;
    private String endpoint = "http://alert-gateway:8080";
    private int timeoutMs = 500;
    // getter / setter 省略
}

// 2. 自动配置类（Boot 3 用 @AutoConfiguration）
@AutoConfiguration
@ConditionalOnClass(AlertChannelClient.class)                 // 依赖存在才装配
@ConditionalOnProperty(prefix = "risk.alert", name = "enabled", havingValue = "true", matchIfMissing = true)
@EnableConfigurationProperties(RiskAlertProperties.class)
public class RiskAlertAutoConfiguration {

    @Bean
    @ConditionalOnMissingBean                                  // 允许业务方自定义覆盖
    public AlertChannelClient alertChannelClient(RiskAlertProperties props) {
        return AlertChannelClient.builder()
                .endpoint(props.getEndpoint())
                .timeoutMs(props.getTimeoutMs())
                .build();
    }
}
```

```
# 3. 文件：src/main/resources/META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports
com.hongwang.risk.alert.RiskAlertAutoConfiguration
```

**面试加分项**：
- 能准确区分「自动配置的 SPI 文件（`.imports`）」和「`SpringFactoriesLoader` 的 `spring.factories`（Boot 3 仍用于 Listener/Initializer）」——说明你读过 Boot 3 的变更，不是照搬旧博客。
- 能说出 `@AutoConfiguration` 相比 `@Configuration` 多了 `proxyBeanMethods = false`（**lite 模式，跳过 CGLIB 增强，启动更快**）和「只能通过 imports 文件注册」的约束。
- 能提「自动配置类加载顺序由 `@AutoConfigureBefore` / `@AutoConfigureAfter` / `@AutoConfigureOrder` 控制」，比如数据源自动配置要在 MyBatis 自动配置之前。
- 结合 Q9：**自动配置太多是启动慢的原因之一**，能通过 `spring.autoconfigure.exclude` 裁剪——把 Q8 和 Q9 串起来答是加分点。

---

### <a id="q9"></a>题 9：SpringBoot 的启动流程是怎样的？线上怎么优化启动速度？

- **题干**：「SpringBoot 项目从 `main` 方法开始，到能接收请求，中间发生了什么？你们线上服务启动多快，如果启动慢你会怎么优化？配置文件加载的优先级也顺便说说。」
- **考察点**：启动主流程 + 工程优化能力（不是只会背 `SpringApplication.run`）+ 配置优先级这种实操细节 + 顺带考跨域/全局异常这类日常问题。

**标准答案**：

**SpringBoot 启动流程（`SpringApplication.run()` 关键步骤）**：
1. **`new SpringApplication(...)`（构造函数阶段）**：
   - 推断应用类型 `WebApplicationType`（有 `DispatcherServlet` 且无 `WebFlux` → `SERVLET`；有 WebFlux → `REACTIVE`；都没有 → `NONE`）；
   - 用 `SpringFactoriesLoader` 从 `spring.factories` 加载 **`ApplicationContextInitializer`** 和 **`ApplicationListener`**（Boot 3 的 `SpringApplication` 构造里走 `getSpringFactoriesInstances`）；
   - 推断主启动类（从 `stackTrace` 找 `main` 方法所在的类）。
2. **`run()` 开始**：启动 `StopWatch` 计时；创建 **`BootstrapContext`**（Boot 3.x，用于早期访问环境）。
3. **发布 `ApplicationStartingEvent`**（最早的监听事件）。
4. **`prepareEnvironment()`**：创建并配置 `Environment`，**加载配置文件**（`application.yml`、命令行参数、`application-{profile}` 等），发布 `ApplicationEnvironmentPreparedEvent`；（`ConfigDataEnvironmentPostProcessor` 负责把配置源加入环境）。
5. **打印 Banner**（`Banner` 接口，可 `spring.main.banner-mode=off` 关掉）。
6. **`createApplicationContext()`**：根据应用类型创建容器，Servlet 场景是 **`AnnotationConfigServletWebServerApplicationContext`**。
7. **`prepareContext()`**：设置环境、执行 `ApplicationContextInitializer`、注册主配置类（`refresh()` 前先把主类注册为 BeanDefinition）、发布 `ApplicationContextInitializedEvent` / `ApplicationPreparedEvent`。
8. **`refreshContext()`** → 调 `AbstractApplicationContext.refresh()`（**容器启动的核心 12 步**：`prepareRefresh` → `obtainFreshBeanFactory` → `prepareBeanFactory` → `postProcessBeanFactory` → `invokeBeanFactoryPostProcessors` → `registerBeanPostProcessors` → `initMessageSource` → `initApplicationEventMulticaster` → `onRefresh` → `registerListeners` → `finishBeanFactoryInitialization`（**实例化所有非懒加载单例**）→ `finishRefresh`）。
   - **内嵌 Tomcat 的启动就在 `onRefresh()` 里**：`ServletWebServerApplicationContext.createWebServer()` → `TomcatServletWebServerFactory.getWebServer()` 创建并启动 Tomcat，把 `DispatcherServlet` 注册进去。**所以 Tomcat 是 SpringBoot 自己 new 出来、跑在同一个 JVM 里的**，不需要外部容器。
9. **`afterRefresh()`**：`callRunners()` 执行 **`ApplicationRunner` / `CommandLineRunner`**。
10. **发布 `ApplicationStartedEvent` / `ApplicationReadyEvent`**（`ready` 表示服务已可对外提供服务，是「预热」的正确时机，见 Q1）。

**图 4-12：SpringApplication.run 的启动流程**

```mermaid
flowchart TD
    A["new SpringApplication 推断 WebApplicationType<br/>加载 ApplicationContextInitializer 与 ApplicationListener"] --> B["run 开始 启动 StopWatch 创建 BootstrapContext"]
    B --> C["发布 ApplicationStartingEvent"]
    C --> D["prepareEnvironment 加载配置文件 application.yml 与 profile<br/>发布 ApplicationEnvironmentPreparedEvent"]
    D --> E["打印 Banner 可用 banner-mode 关闭"]
    E --> F["createApplicationContext Servlet 场景为 AnnotationConfigServletWebServerApplicationContext"]
    F --> G["prepareContext 执行 Initializer 注册主配置类<br/>发布 ApplicationContextInitializedEvent 与 ApplicationPreparedEvent"]
    G --> H["refreshContext 调用 refresh 的 12 步<br/>内嵌 Tomcat 在 onRefresh 里创建并启动"]
    H --> I["afterRefresh 执行 ApplicationRunner 与 CommandLineRunner"]
    I --> J["发布 ApplicationStartedEvent 与 ApplicationReadyEvent<br/>服务真正可用 是预热与注册发现的正确时机"]
```

**启动优化**（按收益排序）：
1. **延迟初始化**：`spring.main.lazy-initialization=true`。默认 `finishBeanFactoryInitialization` 会创建所有单例，懒加载改成「用到才建」。**代价**：首次请求变慢、启动期错误延后暴露——**低峰期/灰度环境可用，核心交易服务慎用**。
2. **裁剪自动配置**：`spring.autoconfigure.exclude=...` 排除不需要的（如没用到 Redis 就排除 `RedisAutoConfiguration`），或用 `@SpringBootApplication(exclude = ...)`。配合 `/actuator/conditions` 看哪些自动配置是「白装配」的。
3. **缩小扫描范围**：`@ComponentScan` 精确指定包（别扫 `com`），减少 `@Configuration`/`@Component` 的解析。
4. **JVM 参数**：`-XX:TieredStopAtLevel=1`（只做 C1 编译，启动快）、`-Xshare:auto`/CDS（Boot 3.2 支持 `spring.aot` + CDS，用 AppCDS 减少类加载）、`-XX:+UseSerialGC`（小堆启动更快）。
5. **AOT / Native Image**：Boot 3 支持 AOT 处理（`mvn spring-boot:process-aot` 生成元数据）+ GraalVM Native Image，启动可到**百毫秒级**，但构建复杂、反射要配 `@RegisterReflectionForBinding`、生态兼容性是坑。
6. **减少启动期重活**：把 `@PostConstruct` 里的重逻辑（远程调用、预热）移到 `ApplicationReadyEvent`（Q1），或用 `spring.main.lazy-initialization` 分包控制。
7. **分层构建**：`spring-boot-maven-plugin` 的 layers，Docker 构建只重传变化的层（配合 Q9 的容器化，发布提速）。

**配置文件优先级（由高到低，面试常考）**：
```
命令行参数 --server.port=9090
  > SPRING_APPLICATION_JSON（环境变量里的 JSON）
  > ServletConfig / ServletContext 初始化参数
  > JNDI
  > Java 系统属性 -Dxxx
  > 操作系统环境变量
  > application-{profile}.properties（jar 包外，./config/ 目录下优先）
  > application-{profile}.properties（jar 包外，当前目录）
  > application-{profile}.properties（jar 包内 classpath）
  > application.properties（依次：外部 config/ > 外部当前目录 > 内部 classpath）
```
关键结论：**外部 > 内部；带 profile 的 > 不带 profile 的；越具体的（更多 profile）越优先**。

**SpringBoot 解决跨域**（三种）：
1. 注解级：`@CrossOrigin`（Controller 类/方法）；
2. 全局（SpringMVC）：`WebMvcConfigurer#addCorsMappings` 配置 `CorsRegistry`；
3. **网关级（微服务推荐）**：在 Spring Cloud Gateway 配 `spring.cloud.gateway.globalcors`，统一在入口解决，各微服务不需要感知跨域。
⚠️ 注意：**`addCorsMappings` 是 WebMVC 的，Gateway 是 WebFlux 的，两者不通用**——在网关里写 WebMvc 配置不会生效（经典踩坑）。

**全局异常处理**：
```java
@RestControllerAdvice                 // = @ControllerAdvice + @ResponseBody
public class GlobalExceptionHandler {
    @ExceptionHandler(BizException.class)
    public Result<Void> handleBiz(BizException e) { return Result.fail(e.getCode(), e.getMessage()); }

    @ExceptionHandler(MethodArgumentNotValidException.class)     // @Valid 校验失败
    public Result<Void> handleValid(MethodArgumentNotValidException e) { /* 取 BindingResult 拼错误信息 */ }

    @ExceptionHandler(Exception.class)                            // 兜底
    public Result<Void> handleUnknown(Exception e) { return Result.fail("SYSTEM_ERROR", "系统繁忙"); }
}
```
- `@ControllerAdvice` 三大用途：**全局异常**（`@ExceptionHandler`）、**全局数据绑定**（`@InitBinder`）、**全局数据预处理**（`@ModelAttribute`）。

**企业实际处理**：
- 核心服务（期权风控）**不用全局懒加载**，因为首笔预警请求不能因为初始化慢而超时；改用「`ApplicationReadyEvent` 预热 + 精准裁剪自动配置」。
- 用 `spring-boot-maven-plugin` 的 `layers` + Docker 多阶段构建：依赖层、快照层、应用层分开，改代码只重传应用层，配合 Jenkins 流水线把发布从 4 小时压到 15 分钟（简历业绩）。
- 全局异常统一返回 `Result`，**但绝不在兜底 `Exception` 里把堆栈返回给前端**（信息泄露），只记日志并返回统一错误码。

**为什么这样做**：
- **因为 `refresh()` 的 12 步是「先注册后实例化」的两阶段设计**：前 11 步都在改 BeanDefinition、注册后置处理器，最后一步 `finishBeanFactoryInitialization` 才真正 new 对象。这样「BeanPostProcessor/BeanFactoryPostProcessor 能影响所有 Bean」——**必须在实例化之前注册好处理器**，否则后注册的处理器对已创建的 Bean 无效。
- **因为内嵌 Tomcat 让「打包即运行」成为可能**：传统 war 部署要装外部 Tomcat、配 context、拷 war 包；内嵌后一个 `java -jar` 就起服务，**这是容器化（Docker）和 DevOps 的前提**。代价是每个应用自带一个 Tomcat（内存开销），但对微服务是可接受的权衡。
- **因为启动优化本质是「延迟 + 裁剪」两招**：启动慢源于「做了太多启动期不需要的事」（创建所有 Bean、加载所有自动配置）。要么**延迟到用时再做**（懒加载、移到 ReadyEvent），要么**干脆不做**（裁剪自动配置）。理解这条主线，就不会死记优化清单。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：精准裁剪自动配置 + ReadyEvent 预热 + 全局异常兜底

```java
// 1. 精准裁剪：只装配用得到的自动配置
@SpringBootApplication(exclude = {
        org.springframework.boot.autoconfigure.jdbc.DataSourceAutoConfiguration.class,
        org.springframework.boot.autoconfigure.mail.MailSenderAutoConfiguration.class
})
public class RiskAlertApplication {
    public static void main(String[] args) {
        SpringApplication.run(RiskAlertApplication.class, args);
    }
}
```

```yaml
# 2. application.yml 关键项
spring:
  main:
    # 核心服务不开懒加载，保证启动即可服务
    lazy-initialization: false
    banner-mode: off
  autoconfigure:
    exclude:
      - org.springframework.boot.autoconfigure.mail.MailSenderAutoConfiguration
# 3. 跨域（网关层统一处理更优，此处为单服务兜底）
  cloud:
    gateway:
      globalcors:
        cors-configurations:
          '[/**]':
            allowedOrigins: "https://risk.hongwang.com"
            allowedMethods: [GET, POST, PUT, DELETE]
            allowCredentials: true
```

**面试加分项**：
- 能区分「`ApplicationContextInitializer`（容器 refresh 前执行，可改容器）」和「`ApplicationRunner`/`CommandLineRunner`（refresh 完成后执行，做启动后任务）」。
- 能说出「`ApplicationReadyEvent` 才是「服务真正可用」的信号，做预热、注册到服务发现、开始消费 MQ 都应该在这之后」——很多人错在 `@PostConstruct` 里做这些。
- 提到 Boot 3.2 的 **`spring.aot`** 和虚拟线程（`spring.threads.virtual.enabled=true`，JDK 21+）——说明你关注新版本。
- 把配置优先级答成「**外部覆盖内部、profile 覆盖默认、命令行覆盖一切**」的口诀，比死记 17 行清单实用。

---

### <a id="q10"></a>题 10：MyBatis 的执行流程，`#{}` 和 `${}` 有什么区别？

- **题干**：「说说 MyBatis 一次查询的完整执行流程，四大核心组件是什么？`#{}` 和 `${}` 的区别？MyBatis 的一级二级缓存说说，还有——Mapper 接口没有实现类，它是怎么被调用的？分页插件什么原理？」
- **考察点**：MyBatis 全链路理解（从 Mapper 代理到结果映射）+ SQL 注入安全意识 + 缓存机制 + 插件原理。一题打包 MyBatis 主线，是中小厂必考。

**标准答案**：

**一、执行流程（一次查询）**：
1. **启动解析**：`SqlSessionFactoryBuilder` 解析 `mybatis-config.xml` + 所有 Mapper XML（或被 `@MapperScan` 扫描的接口），构建全局唯一 **`Configuration`** 对象，`SqlSessionFactory` 由它创建。
2. **获取 SqlSession**：`SqlSessionFactory.openSession()` 得到 `SqlSession`（本质是 `DefaultSqlSession`，内部持有 `Executor`）。
3. **获取 Mapper 代理**：`sqlSession.getMapper(UserMapper.class)` → `MapperRegistry.getMapper()` → **`MapperProxyFactory` 用 JDK 动态代理生成 `MapperProxy`**。
4. **调用 Mapper 方法**：`MapperProxy.invoke()` 判断是 Object 方法还是 Mapper 方法 → 找到对应的 **`MapperMethod`** → 执行 `execute()`。
5. **Executor 执行**：`Executor`（默认 `SimpleExecutor`，经过插件链包装成 `CachingExecutor` 等）：
   - 查缓存 → 未命中则 `queryFromDatabase()`；
   - `Configuration.getMappedStatement(id)` 拿到 **`MappedStatement`**（封装了 SQL、参数映射、结果映射、缓存配置）。
6. **StatementHandler 创建 Statement**：`PreparedStatementHandler` 用 `Connection.prepareStatement(sql)`。
7. **ParameterHandler 设参**：把 Java 参数按 `#{}` 的映射设置到 `PreparedStatement`（`ps.setXxx`）。
8. **执行 SQL**，得到 `ResultSet`。
9. **ResultSetHandler 映射结果**：按 `<resultMap>` / 驼峰映射把结果集转成 Java 对象（含 `TypeHandler` 做 JDBC 类型 ↔ Java 类型转换）。
10. **（插件在此介入）**：`Executor`、`StatementHandler`、`ParameterHandler`、`ResultSetHandler` **四大对象都可以被 `Interceptor`（插件）包装**（`Plugin.wrap` 生成代理）。

**四（五）大核心组件**：
- **`Executor`**：执行器，负责整体调度、缓存、事务（`SimpleExecutor` / `ReuseExecutor` / `BatchExecutor` / `CachingExecutor`）；
- **`StatementHandler`**：负责创建 `Statement`、执行 SQL；
- **`ParameterHandler`**：负责参数设置；
- **`ResultSetHandler`**：负责结果集映射；
- （`MappedStatement`：不是「处理器」但很重要，是 XML 里一个 `<select>` 的完整描述）。

**图 4-13：MyBatis 的执行流程**

```mermaid
flowchart LR
    M["Mapper 接口 无实现类"] --> MP["MapperProxyFactory 生成 JDK 动态代理 MapperProxy"]
    MP --> MM["MapperMethod 找到对应的 MappedStatement"]
    MM --> SS["SqlSession 内部持有 Executor"]
    SS --> EX["Executor 查缓存 未命中则查库"]
    EX --> SH["StatementHandler 创建 PreparedStatement"]
    SH --> PH["ParameterHandler 按预编译占位符设参"]
    PH --> RUN["执行 SQL 得到 ResultSet"]
    RUN --> RH["ResultSetHandler 配合 TypeHandler 映射成 Java 对象"]
```

**二、`#{}` vs `${}`（高频，安全题）**

| 维度 | `#{}` | `${}` |
|------|-------|-------|
| 实现 | **预编译占位符 `?`**（`PreparedStatement`） | **字符串直接拼接**进 SQL |
| SQL 注入 | **安全**（参数值不参与 SQL 解析） | **有注入风险** |
| 类型处理 | 自动按 `TypeHandler` 处理（加引号、转类型） | 原样拼接，需自己保证格式 |
| 用途 | 所有**参数值**（where 条件、insert 值） | **动态 SQL 片段**：#{} 做不到的场景，如动态**表名/列名/`ORDER BY` 字段** |

- **结论：能用 `#{}` 就用 `#{}`；必须用 `${}` 时（表名、order by），参数一定要走白名单校验**。
- 举例：`ORDER BY ${sortField}` 若 `sortField` 来自前端，攻击者可传 `id; DROP TABLE ...`；正确做法是 `sortField` 与合法列名做白名单比对，不合法就拒绝。

**三、缓存**

| 级别 | 范围 | 默认 | 失效条件 |
|------|------|------|----------|
| 一级缓存 | **`SqlSession` 级** | 开启 | 同一个 `SqlSession` 内相同 `statement + 参数 + 分页` 命中；执行 `update/insert/delete`、`commit`、`rollback`、`close` 后**清空** |
| 二级缓存 | **namespace（Mapper）级** | **关闭**（需 `<cache/>` 或 `@CacheNamespace` 开启） | 跨 `SqlSession` 共享；**该 namespace 内有写操作**即清空；实体需 `implements Serializable` |

- 一级缓存是 `SqlSession` 内的 `PerpetualCache`（HashMap），本质是**会话级**；
- 二级缓存是 `TransactionalCacheManager` 管理、`CachingExecutor` 触发，跨会话但**同一 namespace**；
- **坑**：二级缓存跨 `SqlSession` 共享，如果多表关联、其他 namespace 改了数据，本 namespace 缓存不会失效 → **脏读**。所以生产上**微服务场景一般不推荐开二级缓存**（用 Redis 做分布式缓存更可控）。

**四、Mapper 接口为什么不用实现类（动态代理）**
- `@MapperScan` / `@Mapper` 让 MyBatis 启动时扫描接口，注册到 **`MapperRegistry`**（`knownMappers` 是个 `Map<Class, MapperProxyFactory>`）。
- 调用 `getMapper()` 时，`MapperProxyFactory.newInstance()` 用 **`Proxy.newProxyInstance`** 生成 JDK 动态代理 **`MapperProxy`**（实现 `InvocationHandler`）。
- 调用接口方法时进 `MapperProxy.invoke()`：**Object 的 `toString/equals/hashCode` 直接走，默认方法走 `invokeDefaultMethod`，其余走 `MapperMethod.execute()`**——`MapperMethod` 里根据方法全限定名（`namespace + "." + methodName`）找到 `MappedStatement`，再根据 SQL 类型（`SELECT/INSERT/UPDATE/DELETE`）调 `Executor` 对应方法。
- **本质**：接口 → 全限定名 → `MappedStatement` 的映射，是「**约定优于配置**」在 ORM 层的体现。

**五、分页插件原理（PageHelper / MyBatis-Plus PaginationInnerInterceptor）**
- 基于 **`Interceptor` + `Plugin`（JDK 代理）**：通过 **`@Intercepts` / `@Signature`** 声明拦截 `Executor.query`（PageHelper 拦 `Executor#query`，MyBatis-Plus 3.5+ 拦 `Executor` 的 `query`）。
- `intercept()` 里：
  1. 从参数（PageHelper 从 **`ThreadLocal`** `PageMethod`，MP 从 `IPage` 参数）判断是否分页请求；
  2. 先执行一条 **`SELECT COUNT(*)`** 拿总数（可关闭 count 查询）；
  3. 用 `Dialect`（方言：MySQL/H2/Oracle）**改写原 SQL**，追加 `LIMIT ?, ?`；
  4. 继续 `invocation.proceed()` 执行改写后的 SQL。
- **物理分页**（改写 SQL 让数据库只返回一页），不是内存分页。
- **坑**：PageHelper 的 `ThreadLocal` 在**异步/线程池**里会串页——分页参数没清干净，导致「上一个请求的页码影响下一个请求」。用 `PageHelper.clearPage()` 或改用 MyBatis-Plus（基于方法参数，无 ThreadLocal 污染）。

**企业实际处理**：
- **禁用 `${}` 拼参数**，静态扫描（如 SpotBugs/自定义规则）拦截 `${` 出现在 XML 参数位置；必须动态排序时用白名单。
- **二级缓存默认不开**，缓存统一走 Redis（可控 TTL、可控失效、能跨语言）；一级缓存默认开启但不显式依赖（会话很短）。
- 分页统一用 **MyBatis-Plus 的 `IPage`**（无 ThreadLocal 坑），并限制最大页大小（防 `LIMIT 1000000, 10` 拖垮 DB）。
- 复杂查询用 XML 写（可读性好），简单 CRUD 用 MyBatis-Plus 的 `BaseMapper` 省代码。

**为什么这样做**：
- **因为「`#{}` 走 PreparedStatement」是数据库层面最可靠的防注入手段**：预编译让参数值**永远不被当作 SQL 语法解析**，无论参数里有单引号还是 `--` 注释。`${}` 拼接等于把 SQL 结构和数据混在一起，注入就不可避免。
- **因为二级缓存用 namespace 做隔离是「粒度粗」的设计**：namespace 级失效无法感知「别的 Mapper 改了同样一张表」，所以多表关联必然脏读。这是**简单换正确性**的取舍，生产上不如用独立缓存层（Redis）精确控制。
- **因为 Mapper 用动态代理而不是代码生成/字节码增强**：动态代理**零编译期产物**（不需要像 JPA 那样生成实现），接口只写声明，SQL 放 XML，**职责清爽**；代价是运行期不易调试（要看 `MapperProxy` 链路）。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：安全的分页 + 动态排序（白名单）

```java
@Mapper
public interface RiskAlertMapper {

    // #{} 预编译，防注入
    List<AlertRecord> selectBySymbol(@Param("symbol") String symbol,
                                     @Param("status") String status,
                                     @Param("orderColumn") String orderColumn,   // 走白名单校验后再拼
                                     @Param("offset") int offset,
                                     @Param("size") int size);
}
```

```xml
<select id="selectBySymbol" resultType="com.hongwang.risk.entity.AlertRecord">
    SELECT id, symbol, status, triggered_at
    FROM risk_alert_record
    WHERE symbol = #{symbol}                       <!-- 值：预编译，安全 -->
      AND status = #{status}
    ORDER BY ${orderColumn}                        <!-- 动态列：必须白名单 -->
    LIMIT #{offset}, #{size}                       <!-- 值：预编译，安全 -->
</select>
```

```java
@Service
public class AlertQueryService {
    // 动态排序列白名单，杜绝 ${} 注入
    private static final Set<String> ALLOWED_SORT = Set.of("triggered_at", "symbol", "status");

    private final RiskAlertMapper mapper;

    public AlertQueryService(RiskAlertMapper mapper) { this.mapper = mapper; }

    public List<AlertRecord> query(String symbol, String status, String sort, int page, int size) {
        String orderColumn = ALLOWED_SORT.contains(sort) ? sort : "triggered_at";   // 不合法则用默认
        int safeSize = Math.min(size, 100);                                          // 限制最大页大小
        return mapper.selectBySymbol(symbol, status, orderColumn, (page - 1) * safeSize, safeSize);
    }
}
```

**面试加分项**：
- 能说出 `MapperProxy.invoke()` 对 **`Object` 方法（toString/equals/hashCode）和 `default` 方法**的特殊处理——这是读源码的细节。
- 能提「`StatementHandler` 有 `SimpleStatementHandler`（Statement，用于 `${}` 或静态 SQL）、`PreparedStatementHandler`（PreparedStatement，用于 `#{}`）、`CallableStatementHandler`（存储过程）」。
- 能说出「插件用 `@Intercepts` 声明拦截点，`Plugin.wrap()` 生成代理，可嵌套多层（责任链）」，并把「分页插件」和「慢 SQL 插件」「多租户插件」归为同一机制。
- 提到 MyBatis-Plus 的取舍：「省了单表 CRUD 的 SQL，但**复杂 join/多表统计还是得写 XML**；且它的 `LambdaQueryWrapper` 有轻微反射开销，极端 QPS 下单表查询仍是手写 SQL 更快」。

---

## 高阶架构题（5 道）

### <a id="q11"></a>题 11：微服务拆分后 OpenFeign 调用怎么设置超时和熔断降级？

- **题干**：「你们期权风控系统拆成了多个微服务，服务之间用 OpenFeign 调。服务 A 调服务 B 超时了怎么办？如果 B 挂了，A 会不会被拖死？你怎么配超时、重试和熔断降级？说说 OpenFeign 的底层原理。」
- **考察点**：微服务调用治理的实战能力（超时/重试/熔断/隔离）+ 能否讲清 OpenFeign 的代理与负载均衡机制 + 雪崩防护意识。**这题与「Kafka/CAP」模块的可用性主题呼应，但落点在 Spring Cloud 组件使用。**

**标准答案**：

**一、OpenFeign 原理**
1. `@EnableFeignClients` 扫描所有 `@FeignClient` 接口，为每个接口注册一个 **`FeignClientFactoryBean`**。
2. 注入时，`FeignClientFactoryBean.getObject()` 返回 **`Feign` 构建的动态代理**（`ReflectiveFeign.newInstance()`，底层 JDK 动态代理），每个方法对应一个 `MethodHandler`。
3. 调用时：
   - `MethodHandler` 根据 `@RequestMapping` 注解生成 **`RequestTemplate`**（URL、方法、Header、Body）；
   - 交给 **`Client`** 发起 HTTP 请求（默认 `feign.Client.Default` 走 `HttpURLConnection`，可换 **OkHttp / Apache HttpClient 5**）；
   - 请求前经 **`LoadBalancerClient`**（Spring Cloud LoadBalancer）**按服务名选实例**（从注册中心拿到的实例列表里按轮询/随机选）；
   - 经过 `RequestInterceptor`（可加 TraceId、鉴权头）、`Decoder`/`Encoder`（JSON 编解码）、`ErrorDecoder`（错误处理）。
4. 集成 Sentinel 后，Feign 调用被 Sentinel 资源包裹，可做熔断降级。

**图 4-14：微服务调用链与熔断降级**

```mermaid
flowchart LR
    GW["Spring Cloud Gateway<br/>鉴权 限流 TraceId"] --> RE["规则引擎服务"]
    RE -->|OpenFeign 调用| AG["预警生成服务"]
    AG -->|OpenFeign 调用| AC["告警通道服务"]
    AG -.->|熔断降级| SENT["Sentinel 熔断器<br/>CLOSED OPEN HALF_OPEN"]
    SENT -->|OPEN 直接走 fallback| FB["fallbackFactory 降级<br/>落库待补发 绝不静默丢弃"]
    RE -.->|超时递减 下游小于上游| AG
    AG -.->|TraceId 通过请求头透传| AC
```

**二、超时配置（必须显式配，默认值会坑人）**
```yaml
spring:
  cloud:
    openfeign:
      client:
        config:
          default:                       # 对所有 FeignClient 生效
            connectTimeout: 1000         # 连接超时 1s
            readTimeout: 2000            # 读超时 2s
          alert-service:                 # 针对指定服务覆盖
            readTimeout: 500             # 告警通道要求快失败，500ms
            loggerLevel: basic
```
- **默认 `readTimeout` 是 60s**（Feign 默认 10s，但 Spring Cloud 早期有 60s 的坑），必须显式设小；风控实时预警链路要求**「快速失败」而非「长时间等待」**，所以超时要短（500ms~2s）。
- 注意：**LoadBalancer 层也有超时**（`spring.cloud.loadbalancer.retry.*`），Feign 层和 LB 层的超时要一起看，否则上层以为超时了下层还在重试。

**三、重试**
- Feign 自带 `Retryer`，**Spring Cloud OpenFeign 默认 `Retryer.NEVER_RETRY`（不重试）**，要看 `spring.cloud.openfeign.client.config.default.retryer`。
- 重试的两个前提：**① 接口幂等**（GET 天然幂等，POST 要保证幂等键）；**② 有重试上限 + 退避**，否则雪崩时重试会放大流量。
- **风控预警场景慎用自动重试**：预警触发是**写操作**，重试可能导致重复告警；若必须重试，用「幂等键 + 去重表」。

**四、熔断降级（Sentinel）**
- 接入：`feign.sentinel.enabled=true`（旧版）或 Spring Cloud Alibaba 新版配置；`@FeignClient(name = "alert-service", fallback = AlertClientFallback.class)` 或 `fallbackFactory`（能拿到异常原因）。
- Sentinel 熔断降级策略：
  - **慢调用比例**：RT > 阈值且比例超阈值 → 熔断；
  - **异常比例**：异常占比超阈值 → 熔断；
  - **异常数**：异常数超阈值 → 熔断。
- 熔断状态机：**CLOSED（正常）→ OPEN（熔断，直接走 fallback）→ HALF_OPEN（放一个请求试探，成功则恢复 CLOSED，失败继续 OPEN）**。
- **降级策略要贴合业务**：告警通道挂了，降级成「写本地 MQ 稍后重试」或「落库待补发」，**绝不能静默丢弃预警**（监管级系统丢预警是重大事故）。

**五、限流（Sentinel 流控）**
- **QPS 流控 / 线程数流控**；流控模式：**直接、关联（A 关联 B，B 达阈值限 A）、链路（只限某入口）**；
- 流控效果：**快速失败、Warm Up（预热，冷启动逐步放量）、排队等待（匀速排队，漏桶）**；
- **热点参数限流** `@SentinelResource` + `@SentinelResource(value="...", blockHandler="...")`，对某参数值特别限流（如某个 symbol 的规则查询突增）。

**企业实际处理**：
- 风控预警调用链：`rule-engine → alert-gateway → alert-channel`，每一跳都配 `connectTimeout=500ms / readTimeout=1s`，**超时必须小于上游的总超时**，形成「递减超时」避免级联放大。
- `fallbackFactory` 里**降级不抛异常**，返回一个「降级结果对象」并打点上报，让上游能判断「这是降级数据」而非真实结果。
- Sentinel 规则不写死在代码里，而是**通过 Apollo/Nacos 持久化**（Sentinel 规则持久化到配置中心），支持动态调整阈值、灰度。
- 线程隔离：Sentinel 提供**信号量隔离**（默认，轻量）和**线程池隔离**（重，但能隔离慢调用），对第三方（告警短信通道）用线程池隔离，防止其拖垮自身。

**为什么这样做**：
- **因为「超时 + 熔断」是防雪崩的两道闸**：超时防止「一个慢调用占死线程」，熔断防止「持续打挂掉的依赖」。没有这两道闸，服务 B 一挂，服务 A 的线程全被 B 的调用占满，A 也挂，进而拖垮上游——**这就是雪崩**，微服务里最致命的故障模式。
- **因为「递减超时」是避免级联放大的关键**：如果下游超时 > 上游超时，上游早就超时了、下游还在干，白耗资源。让 `下游超时 < 上游超时`，保证上游超时前下游一定先返回（成功或失败），资源不悬空。
- **因为「降级不能丢数据」是风控业务的硬约束**：技术上的「优雅降级」在普通业务里可以是「返回缓存老数据」，但监管级预警系统里**丢一条预警可能就是合规事故**，所以降级必须落到「持久化 + 补偿」而非「丢弃」。这是**业务约束反过来决定技术方案**的典型。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：OpenFeign + Sentinel fallbackFactory 降级到「落库待补发」

```java
@FeignClient(name = "alert-channel-service",
        fallbackFactory = AlertChannelFallbackFactory.class)          // 用 Factory 能拿到异常原因
public interface AlertChannelClient {

    @PostMapping("/channel/send")
    Result<Void> send(@RequestBody AlertMessage message);
}
```

```java
@Component
public class AlertChannelFallbackFactory implements FallbackFactory<AlertChannelClient> {

    private static final Logger log = LoggerFactory.getLogger(AlertChannelFallbackFactory.class);
    private final PendingAlertRepository pendingRepository;    // 降级落库，稍后补偿

    public AlertChannelFallbackFactory(PendingAlertRepository pendingRepository) {
        this.pendingRepository = pendingRepository;
    }

    @Override
    public AlertChannelClient create(Throwable cause) {
        return message -> {
            log.warn("告警通道降级，落库待补发 messageId={}", message.id(), cause);
            // 绝不能静默丢弃：落库 + 由补偿任务重试，保证预警不丢
            pendingRepository.save(PendingAlert.from(message, cause.getMessage()));
            return Result.fail("DEGRADED", "通道降级，已落库待补发");
        };
    }
}
```

**面试加分项**：
- 能主动讲「**超时要小于上游超时，形成递减**」，说明你处理过真实调用链，不是背概念。
- 能区分 Sentinel 的「**熔断降级**」（针对不稳定依赖）和「**限流**」（保护自身不被流量打垮）——很多人混为一谈。
- 能提到 OpenFeign 换 `Client` 为 **OkHttp/Apache HttpClient 5** 的原因（连接池复用，默认 `HttpURLConnection` 无连接池，高并发下性能差），这是实操细节。
- 提到「`ErrorDecoder` 可以把下游的非 2xx 响应转成自定义异常，配合 Sentinel 的异常统计做熔断」——把 Feign 和 Sentinel 串起来。
- 结合 CAP：熔断降级本质是**牺牲一致性/功能换取可用性（A 优先）**，与「CAP/BASE」模块呼应。

---

### <a id="q12"></a>题 12：Apollo 配置中心原理，风控规则改完怎么做到秒级甚至毫秒级生效？

- **题干**：「你们用 Apollo 做配置中心，说说 Apollo 的原理——它是推还是拉？配置改完怎么通知到客户端？@RefreshScope 是怎么动态刷新的？另外结合实际——风控规则改了，怎么保证在预警服务里快速生效，甚至毫秒级？」
- **考察点**：配置中心机制（推拉结合、长轮询）+ 动态刷新的实现原理（Spring 集成）+ 能否结合业务设计「配置生效」链路。**注意：不要讲注册中心选型（那是另一模块），此题只讲 Apollo 的使用与原理。**

**标准答案**：

**一、Apollo 架构与「推拉结合」**
- 角色：**Config Service**（配置读取 + 推送）、**Admin Service**（配置管理 + 发布）、**Config DB / Portal DB**（MySQL 存储配置与发布历史）、**Portal**（管理界面）、**Client**（各应用里的 `apollo-client`）。
- **推拉结合**（核心机制）：
  1. **客户端定时拉取（pull）**：`RemoteConfigRepository` 启动定时任务，默认**每 5 分钟全量拉一次**（兜底）。同时本地有**缓存文件**（`/opt/data/{appId}/config-cache/...`），启动时先读缓存，保证 Config Service 挂了也能启动。
  2. **长轮询（long polling）实现准实时推送**：客户端发起 `notifications/v2` 请求，**Config Service 把请求 hold 住最多 30 秒**（用 `DeferredResult` + `ScheduledExecutorService`）：
     - 期间若该 namespace 有配置变更（Admin Service 发布后通过 **`ReleaseMessage`** 表 + 定时扫描（默认 1s 扫 `ReleaseMessage` 表）感知），立即返回「有变更」；
     - 若 30 秒无变更，超时返回，客户端**立即再发起**下一轮长轮询。
  3. 客户端收到「有变更」后，**带 `notificationId` 精确拉取**变动的 namespace 的增量配置。
- **效果**：正常情况下配置变更 **1 秒内**（取决于 `ReleaseMessage` 扫描间隔）到达客户端，**是准实时（秒级），不是长轮询的 5 分钟**。

**图 4-15：Apollo 配置热更新的推拉结合机制**

```mermaid
sequenceDiagram
    participant CLI as Apollo Client
    participant CS as Config Service
    participant DB as Config DB
    participant AS as Admin Service
    participant HOLD as RiskRuleHolder

    Note over CLI,DB: 客户端定时全量拉取 默认 5 分钟 作为兜底
    CLI->>CS: 启动时先读本地缓存文件 保证配置中心挂了也能启动
    CLI->>CS: 发起长轮询 notifications/v2
    CS->>CS: 用 DeferredResult 把请求 hold 住最多 30 秒
    AS->>DB: 发布配置写入 ReleaseMessage 表
    CS->>DB: 定时扫描 ReleaseMessage 表 默认 1 秒
    DB-->>CS: 发现该 namespace 有变更
    CS-->>CLI: 立即返回有变更
    CLI->>CS: 带 notificationId 精确拉取增量配置
    CS-->>CLI: 返回新配置
    CLI->>HOLD: 触发 ApolloConfigChangeListener
    HOLD->>HOLD: 解析成不可变 RuleSnapshot 用 AtomicReference 原子替换
    Note over HOLD: 读取永远走内存 切换微秒级 读取纳秒级
```

**二、灰度发布**
- Apollo 支持按 **IP / Label（自定义标签）** 灰度：一个配置项可以有「默认值 + 灰度值 + 灰度规则」。
- 发布时可选择「**全量发布**」或「**灰度发布**」，灰度规则匹配的客户端拿到灰度值，其余拿默认值。
- 用途：**新规则先在小流量机器验证，再全量**——风控规则尤其需要这个（上线错误规则可能造成误报风暴）。
- 灰度后还有「**放弃灰度**」「**全量发布**」两个操作。

**三、`@RefreshScope` 动态刷新原理**
- `@RefreshScope` 的 Bean 被放在一个特殊 Scope（`refresh` scope）里，**实际注入的是它的代理**（`ScopedProxyMode.TARGET_CLASS`）。
- **刷新时不清除 Bean 本身，而是清空 `RefreshScope` 的缓存**（`destroy()` → `this.scopeCache` 清空）；下次访问该 Bean 时，**重新创建**（`get()` → 重建 Bean，重新从 `Environment` 读 `@Value`）。
- **前提**：需要发布 `EnvironmentChangeEvent` 触发刷新——Spring Cloud Context 的 `RefreshEventListener` 监听 `EnvironmentChangeEvent`，调用 `ContextRefresher.refresh()` → `refreshScope.refreshAll()`。
- **Apollo 的 `@Value` 自动更新**：Apollo 的 Spring 集成（`ApolloConfigChangeListener` + `AutoUpdateConfigChangeListener`）会**用反射直接更新 `@Value` 字段**（不依赖 `@RefreshScope` 重建 Bean）——所以 Apollo 场景下 `@Value` 也能自动刷新。**但 `@ConfigurationProperties` 的 Bean 建议加 `@RefreshScope` 或用 Apollo 的 `@ApolloConfigChangeListener` 手动处理**。
- **最佳实践**：用 **`@ApolloConfigChangeListener`** 显式监听，
  ```java
  @ApolloConfigChangeListener("risk-rules")
  public void onChange(ConfigChangeEvent e) { /* 重新加载规则到本地缓存 */ }
  ```

**四、风控规则「毫秒级生效」的落地设计**
Apollo 本身是**秒级**（长轮询 + 1s 扫描），要做到**毫秒级生效**，必须靠**「事件驱动 + 本地缓存 + 无锁读取」**：
1. **规则不落 `@Value`，而是用 `@ApolloConfigChangeListener` 监听变更**，变更时**原子替换本地缓存的规则对象**（用 `AtomicReference<RuleSnapshot>` 直接换引用，无锁）。
2. **规则读取走本地缓存**（Caffeine / `AtomicReference`），**每次预警判断都是内存读取，纳秒级**——不能每次都查 Apollo 或 DB。
3. **Apollo 变更到达（秒级）→ 触发 listener → 原子替换快照（微秒级，无锁）→ 下一个请求立即用新规则**。这样「端到端」是**「Apollo 推送延迟（秒级）+ 替换延迟（微秒）」**，对业务表现为「改完基本立即生效」。
4. **多实例一致性**：Apollo 是每个实例各自长轮询，多个预警实例会在**同一个秒级窗口内**陆续拿到变更；如果要求「所有实例同时切换」（硬约束），需要额外用**版本号 + 双缓冲**：变更时**先加载新规则到影子快照，再用 CAS 切换指针**，保证单实例内「要么全用旧、要么全用新」，避免读到半更新状态。
5. **兜底**：Apollo 挂了，客户端用**本地缓存文件**继续运行（默认 `/opt/data/.../config-cache`），保证「配置中心不可用 ≠ 服务不可用」。

**企业实际处理**：
- `@Value` 只用于**少量、简单**配置；**规则表、灰度开关、限流阈值**用 `@ConfigurationProperties` + 监听器 + 本地缓存。
- 配置项按 **namespace 拆分**：`application`（公共）、`risk-rules`（风控规则）、`alert-channel`（通道配置），各服务只订阅自己关心的 namespace。
- **规则热更新的原子性**：用 `AtomicReference<RuleSnapshot>` 整体替换（不可变对象），**绝不逐条改 Map**——否则并发下会读到「改了一半」的规则集。
- **变更审计**：Apollo 有发布历史（谁在什么时候改了什么），监管场景必须启用（合规要求可追溯）。

**为什么这样做**：
- **因为「推拉结合」是配置中心的通用最优解**：纯推送（服务端维护全部客户端连接）在客户端量大时是灾难（连接管理、广播风暴）；纯拉取（定时轮询）有延迟且浪费请求。**长轮询 = 拉取的形式 + 推送的时效**：连接在服务端 hold 住，有变更立即返回，无变更超时重连——**用最少的资源拿到准实时的效果**。
- **因为「本地缓存 + 原子替换」是毫秒级生效的前提**：Apollo 的网络延迟无法消除（物理链路），但**规则读取必须永远走内存**，把「配置到达」和「规则生效」解耦——到达是秒级（网络决定），生效是微秒级（内存操作决定）。**这是「把延迟放到变更路径而不是读取路径」的经典设计**。
- **因为「不可变快照 + 原子指针切换」保证一致性**：风控规则是一个**规则集**，如果逐条修改，并发线程可能读到「只更新了 3 条规则」的中间态，导致误判。用 `AtomicReference` 换整个不可变快照，保证**读到的永远是某个完整版本**——这是「用不可变对象换线程安全」的取舍。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：风控规则通过 Apollo 监听 + 原子快照实现「准实时生效」

```java
@Component
public class RiskRuleHolder {

    private static final Logger log = LoggerFactory.getLogger(RiskRuleHolder.class);

    // 不可变快照 + 原子引用：读取永远无锁，切换永远原子
    private final AtomicReference<RuleSnapshot> snapshot = new AtomicReference<>(RuleSnapshot.empty());

    private final RuleParser ruleParser;   // 把 Apollo 的 JSON 解析成规则对象

    public RiskRuleHolder(RuleParser ruleParser) {
        this.ruleParser = ruleParser;
        // 1. 启动时先加载一次（Apollo 客户端会自动读本地缓存，保证 Config Service 挂了也能启动）
        refresh("startup");
    }

    // 2. 监听 Apollo 指定 namespace 的变更，变更即刷新
    @ApolloConfigChangeListener("risk-rules")
    public void onRuleChange(ConfigChangeEvent event) {
        if (event.isChanged("rules")) {
            log.info("检测到风控规则变更，开始热更新 changedKeys={}", event.changedKeys());
            refresh("apollo-change");
        }
    }

    private void refresh(String source) {
        String json = ApolloConfig.getConfig("risk-rules").getProperty("rules", "[]");
        RuleSnapshot newSnapshot = ruleParser.parse(json);   // 解析成不可变对象
        snapshot.set(newSnapshot);                            // 原子切换，微秒级，无锁
        log.info("风控规则快照已切换 source={} ruleCount={}", source, newSnapshot.size());
    }

    // 3. 预警主链路：只读内存，纳秒级
    public RuleSnapshot current() {
        return snapshot.get();
    }
}
```

**面试加分项**：
- 能说清「Apollo 的『推送』本质是**长轮询**，不是长连接/WebSocket」——这是很多人误解的点。
- 能区分「Apollo 用**反射直接改 `@Value` 字段**」vs「`@RefreshScope` 用**重建 Bean**」两条不同的刷新路径。
- 能提「Config Service 会把长轮询请求 hold 30s，靠 `ReleaseMessage` 表 + 定时扫描（默认 1s）感知变更」——这个 1s 是「配置变更多久到达」的关键参数，可调。
- 主动区分「**配置中心（Apollo/Nacos config）**」和「**注册中心（Nacos discovery）**」——虽然 Nacos 二合一，但职责不同，这体现架构清晰度（但**不展开选型对比**）。
- 结合监管场景说「发布历史 + 灰度 + 秒级回滚」对风控规则的重要性：**错误规则上线后可秒级回滚**，是风控系统的安全网。

---

### <a id="q13"></a>题 13：Gateway 断言和过滤器怎么工作？TraceId 是怎么跨服务透传的？

- **题干**：「微服务网关你们用的 Spring Cloud Gateway 吧？说说它的路由、断言、过滤器怎么工作？它跟 Zuul 1.x 有什么区别，为什么基于 WebFlux？还有——链路追踪的 TraceId 是怎么从网关一路传到下游服务的？」
- **考察点**：网关核心机制 + 技术选型理解（WebFlux 非阻塞）+ 链路追踪的透传原理（MDC + Header）。**Gateway 属于 Spring Cloud 使用层面，是本题范围；不涉及注册中心选型。**

**标准答案**：

**一、Gateway 三大核心概念**
1. **Route（路由）**：网关的基本单元，`id + uri + predicates + filters`。一个请求匹配到 Route 后，按 filters 处理后转发到 `uri`。
2. **Predicate（断言）**：**路由匹配条件**，`RoutePredicateFactory` 生成。常用的：
   - `Path=/api/risk/**`、`Method=POST`、`Header=X-Token, \d+`、`Query=symbol`、`Cookie=sessionId`、`After/Before/Between`（时间窗口，可做**定时发布/灰度**）。
   - 断言是「与」关系，全满足才匹配。
3. **Filter（过滤器）**：
   - **`GatewayFilter`（路由级）**：只对某个 Route 生效（如 `StripPrefix=1`、`RequestRateLimiter`）；
   - **`GlobalFilter`（全局级）**：对所有 Route 生效（如 `NettyRoutingFilter` 转发、`LoadBalancerClientFilter` 选实例）；
   - 常见内置：`AddRequestHeader`、`RewritePath`、`RequestRateLimiter`（Redis 令牌桶限流）、`CircuitBreaker`（熔断）。
   - **执行顺序**：按 `Ordered` 排序，`pre` 阶段正序执行，`post` 阶段逆序执行（洋葱模型，和 AOP 一致）。

**请求处理流程**：
```
请求 → RoutePredicateHandlerMapping 匹配 Route
     → 组装 FilterChain（GlobalFilter + 该 Route 的 GatewayFilter，按 order 排序）
     → pre 阶段依次执行（改请求：加头、鉴权、限流）
     → NettyRoutingFilter 转发到下游服务（WebClient）
     → post 阶段逆序执行（改响应）
     → 返回给客户端
```

**二、与 Zuul 1.x 的区别 / 为什么基于 WebFlux**

| 维度 | Zuul 1.x | Spring Cloud Gateway |
|------|----------|----------------------|
| 编程模型 | **Servlet 阻塞 IO**，一请求一线程 | **WebFlux + Reactor Netty 异步非阻塞** |
| 并发能力 | 线程数即并发上限，撑不住高并发 | 少量线程（EventLoop）支撑大量并发连接 |
| 过滤器 | `ZuulFilter`（pre/route/post/error） | `GatewayFilter` + `GlobalFilter`，基于 Reactor |
| 长连接/SSE/WebSocket | 支持一般 | 原生异步，更适合 |
| 生态 | Netflix 停更，Zuul 2 未被 Spring Cloud 整合 | Spring 官方主推 |

- **为什么 WebFlux**：网关是**所有流量的入口**，是典型 IO 密集型场景（大量请求 + 等待下游响应）。**阻塞模型下线程都耗在等 IO 上**，加线程又受限于内存和上下文切换；**异步非阻塞用少量 EventLoop 就能扛住高并发**，这正是网关最需要的。
- **代价**：WebFlux 里**不能写阻塞代码**（`Thread.sleep`、阻塞 JDBC 调用会拖垮 EventLoop），所有下游调用必须用响应式 Client；且**WebFlux 与 WebMvc 不能混用**（一个应用只有一个 Web 栈）——**这就是为什么「网关里配 `addCorsMappings` 不生效」，因为那是 WebMvc 的配置**。

**三、TraceId 跨服务透传原理**
1. **入口生成**：请求到达网关（或第一个服务）时，若 Header 里没有 TraceId（如 `traceparent` / `X-B3-TraceId`），**生成一个**；有则沿用（保证上游系统传下来的链路不断）。
2. **写入 MDC**：把 TraceId 放进 **`MDC`（SLF4J 的 Mapped Diagnostic Context，底层是 ThreadLocal）**，日志配置 `%X{traceId}` 就能自动打印。
3. **出站透传（关键）**：
   - 网关转发时，用 `GlobalFilter` 把 TraceId 写进下游请求 Header；
   - 服务内用 **OpenFeign 的 `RequestInterceptor`** 或 **`RestTemplate` 的 `ClientHttpRequestInterceptor`**，**从 MDC 取出 TraceId，塞进出站请求 Header**；
   - 下游服务同样在入口（Filter/Interceptor）从 Header 取出 TraceId，写入自己的 MDC。
4. **形成链路**：每跳都「取 Header → 进 MDC → 出站时写 Header」，TraceId 就串起了整条链路。
5. **技术栈演进**：
   - **Spring Boot 2.x**：Spring Cloud Sleuth（基于 Brave，B3 传播格式 `X-B3-TraceId`）；
   - **Spring Boot 3.x**：Sleuth 停更，官方改用 **Micrometer Tracing**（`micrometer-tracing-bridge-brave` 或 `micrometer-tracing-bridge-otel`），传播格式支持 **W3C Trace Context（`traceparent`）/ B3**；
   - **无侵入方案**：SkyWalking / OpenTelemetry Agent（字节码增强），无需改代码，但依赖 Agent 挂载。

**四、跨线程的坑（高频追问）**
- `MDC` 基于 `ThreadLocal`，**子线程拿不到父线程的 TraceId**（异步、线程池场景会丢）。
- 解决：① 用 `TaskDecorator` 包装线程池，提交任务时**复制 MDC 到子线程**；② 用 `TransmittableThreadLocal`（阿里 TTL）跨线程池传递；③ 响应式场景（WebFlux/Gateway）用 Reactor 的 **`Hooks.onEachOperator` + `contextWrite`** 把 TraceId 放 Reactor Context（不是 ThreadLocal）。

**企业实际处理**：
- 网关统一做：**鉴权（JWT 校验）、限流（Redis 令牌桶）、TraceId 生成、CORS、灰度路由（按 Header/用户标签路由到不同版本）**，下游服务专注业务。
- TraceId 用 **W3C `traceparent`** 标准格式（`00-{traceId}-{spanId}-{flags}`），跨语言/跨厂商兼容性最好。
- 日志规范：`logback` 的 pattern 加 `[%X{traceId}]`，配合 ELK/ Loki，**一个 TraceId 能查出全链路日志**——这是排查线上问题的命脉（风控预警出问题时，从网关 TraceId 一路追到规则计算、到告警发送）。
- 线程池统一用带 `TaskDecorator` 的包装（团队提供的 `MdcThreadPoolTaskExecutor`），从根上杜绝 TraceId 丢失。

**为什么这样做**：
- **因为网关必须用异步非阻塞，是「流量入口」这个位置决定的**：入口的 QPS 是全局最高，阻塞模型的线程数就是并发天花板，横向扩机器成本高。异步模型用少量 EventLoop 扛住连接，**用「编程复杂度」换「吞吐与资源效率」**——这个取舍对网关是划算的。
- **因为 TraceId 必须靠「Header 透传」而非「共享内存」**：跨进程没有共享内存，唯一能带信息的通道就是协议（HTTP Header）。**每跳都要「读进来、存下去、带出去」**，任一环缺失链路就断——所以要用 `RequestInterceptor` / `ClientHttpRequestInterceptor` 统一在「出口」注入，而不是每个业务方法手写。
- **因为 WebFlux 与 WebMvc 不能混用是「一个应用一个 Web 栈」的硬约束**：两者都抢占 `DispatcherServlet` / `DispatcherHandler` 的入口，Spring Boot 会按 classpath 二选一。**这解释了网关里 webmvc 配置失效的根因**——不是配置写错，而是压根没有 WebMvc 栈。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：网关全局 TraceId 过滤器 + Feign 出站透传

```java
// 1. 网关：生成/沿用 TraceId，写入 MDC 与下游 Header
@Component
public class TraceIdGlobalFilter implements GlobalFilter, Ordered {

    private static final String TRACE_ID = "traceId";

    @Override
    public Mono<Void> filter(ServerWebExchange exchange, GatewayFilterChain chain) {
        var request = exchange.getRequest();
        String traceId = request.getHeaders().getFirst(TRACE_ID);
        if (traceId == null || traceId.isBlank()) {
            traceId = UUID.randomUUID().toString().replace("-", "");
        }
        String finalTraceId = traceId;
        // 写进下游请求头，同时改写响应头回传给前端
        ServerWebExchange mutated = exchange.mutate()
                .request(r -> r.header(TRACE_ID, finalTraceId))
                .build();
        mutated.getResponse().getHeaders().add(TRACE_ID, finalTraceId);
        return chain.filter(mutated);
    }

    @Override
    public int getOrder() { return Ordered.HIGHEST_PRECEDENCE; }   // 最先执行，保证后续 filter 都能拿到
}
```

```java
// 2. 服务内：Feign 出站时把 MDC 的 TraceId 塞进请求头
@Configuration
public class FeignTraceConfig {

    @Bean
    public RequestInterceptor traceRequestInterceptor() {
        return template -> {
            String traceId = MDC.get("traceId");
            if (traceId != null) {
                template.header("traceId", traceId);
            }
        };
    }
}
```

**面试加分项**：
- 能说出 Gateway 的 pre/post 是**洋葱模型**（和 AOP 一样），`GlobalFilter` 用 `Ordered` 控制顺序。
- 能提「`RequestRateLimiter` 基于 Redis + Lua 令牌桶做分布式限流，`keyResolver` 决定限流维度（IP/用户/接口）」——网关限流是保护后端的核心手段。
- 能区分「**网关限流**（保护整个后端）」和「**Sentinel 服务内限流**（保护单个服务）」，两者配合使用。
- 提到 W3C `traceparent` 格式的字段含义（version-traceId-spanId-flags），说明你了解标准而非某个实现。
- 提到「异步/线程池会导致 MDC 丢失 TraceId」，并给出 `TaskDecorator` / `TransmittableThreadLocal` 的解法——这是真实踩过的坑。
- 结合 Boot 3 说明 Sleuth 已停更、改用 Micrometer Tracing——版本敏感度加分。

---

### <a id="q14"></a>题 14：分布式任务调度（XXL-Job/SchedulerX）原理，多环境主备怎么切换？

- **题干**：「你们风控系统里有分布式任务调度吧？用的是 XXL-Job 还是 SchedulerX？说说它的原理——任务怎么被触发、怎么保证不重复执行？如果任务机器挂了怎么办？多环境/主备机房怎么用配置切换？」
- **考察点**：调度框架原理（中心化调度 + 注册心跳 + 触发）+ 分布式任务的核心难点（幂等、防重复、失败重试、分片）+ 配置中心在「环境切换」里的实际运用。**这是简历业绩「任务依赖调度」的直接落点。**

**标准答案**：

**一、XXL-Job 架构与原理**
- 两个角色：
  - **调度中心（admin）**：管理任务、维护 cron、触发调度；
  - **执行器（executor）**：内嵌在业务应用里（`xxl-job-executor`），注册到调度中心并接收触发。
- **注册与心跳**：执行器启动后，**向调度中心注册**（写注册表，如 DB/注册中心），并**定时心跳**（默认 30s）上报健康；调度中心维护「在线执行器列表」。执行器下线/心跳超时 → 从列表移除。
- **触发流程**：调度中心线程扫描任务，**cron 到期** → 通过 RPC（默认 HTTP，可换 Netty/gRPC）**下发 `run` 请求给执行器** → 执行器用**业务线程池**执行 JobHandler → 回调上报结果。
- **路由策略**：`FIRST`（第一个）、`ROUND`（轮询）、`RANDOM`、`CONSISTENT_HASH`（一致性哈希，同一任务固定到同一机器）、`FAILOVER`（故障转移）、`SHARDING_BROADCAST`（**分片广播**，所有执行器各分一片）。
- **阻塞处理策略**：`单机串行`、`丢弃后续`、`覆盖之前`——解决「同一任务上一次没跑完、下一次又来了」。
- **失败处理**：失败重试（可配次数）、超时告警、`失败告警邮件`。
- **重要**：**调度本身的高可用**——调度中心可集群部署，依赖 DB 行锁/乐观锁（`xxl_job_lock` 表）保证**同一任务同一时刻只被一个调度中心触发**，避免重复调度。

**二、SchedulerX（阿里云）**
- **Serverless 调度**：阿里云托管，无需自建 admin；
- 能力：**秒级 cron**（`0/5 * * * * ?` 能到秒）、**工作流（DAG，任务依赖编排）**、**分片（MapReduce 式）**、**幂等控制**（`blockType` 阻塞策略）、**任务级灰度**；
- 触发原理类似：客户端（SchedulerX Agent）注册到服务端，服务端按计划触发下发。
- **工作流（DAG）**：多个 Job 组成有向无环图，**上游成功才触发下游**——这正是简历里「**任务依赖调度**」的实现方式（如「数据采集 → 规则计算 → 预警生成」三步依赖）。

**三、核心难点与解法**

| 难点 | 原因 | 解法 |
|------|------|------|
| **重复执行** | 网络抖动 / 调度中心重试 / 多调度中心 | ① 业务层**幂等**（唯一业务键 + 去重表 / `INSERT IGNORE` / Redis `SETNX`）；② 调度中心用**分布式锁**（XXL-Job 的 `xxl_job_lock`）保证单次触发 |
| **任务丢失** | 执行器宕机、下发失败 | 失败重试 + 超时补偿 + **调度日志表**核对 |
| **长任务阻塞** | 上一轮未结束下一轮到来 | 阻塞处理策略（丢弃/串行/覆盖）+ 缩短任务粒度 |
| **单机瓶颈** | 一个任务处理全量数据 | **分片广播**：每个执行器处理 `hash(id) % shardTotal == shardIndex` 的那部分 |
| **时钟漂移** | 多机时间不一致 | 统一 NTP 同步；或由调度中心（而非执行器）判断到期 |

- **幂等是分布式调度的第一性要求**：因为网络不可靠（重试）和集群化（可能重复触发），**任何调度任务都必须假设「可能被重复执行」**，用**唯一业务键 + 状态机**保证「重复执行也不会产生副作用」。风控预警生成任务尤其重要——**重复生成 = 重复告警 = 误报风暴**。

**四、多环境主备切换（结合配置中心）**
- **环境隔离**：Apollo 用**不同环境（DEV/FAT/UAT/PRO）部署独立集群**，或同一集群用**不同 cluster（默认 default / 北京 / 上海）**隔离；客户端启动参数 `env` / `apollo.cluster` 决定读哪套配置。
- **主备双活切换**：
  1. 主备机房各部署一套服务，配置**同 namespace 不同 cluster**（如 `cluster=shanghai` / `cluster=beijing`）；
  2. 一个**开关配置**（如 `risk.primary.region=shanghai`）决定哪边是主；
  3. 切换时**改开关 → Apollo 秒级推送到所有实例 → 各实例按新开关决定「是否接管流量/任务」**；
  4. 关键：**切换必须幂等且带防脑裂**——用**分布式锁（Redis/ZK）选主**保证同一时刻只有一个机房在跑定时任务，否则主备双跑会导致重复。
- **任务调度的主备**：定时任务**只在一个机房启用**（`xxl-job` 执行器只在主机房注册，或 SchedulerX 的任务只在主机房实例上启用），备机房**待命但注册为「禁用」**，切换时启用备机房执行器。

**企业实际处理**：
- 简历里的「**任务依赖调度**」用 **SchedulerX 工作流（DAG）** 或 XXL-Job 的「子任务」串联：采集 Job 成功 → 触发计算 Job → 触发预警 Job，**每步可单独重试、可观测**。
- **所有调度任务强制幂等**：以「业务日期 + 任务类型」或「批次号」为唯一键，重复执行时直接跳过（`INSERT ... ON DUPLICATE KEY UPDATE` 或先查状态）。
- **调度日志 + 告警**：调度失败/超时立刻告警；风控数据传输类任务**双跑校验**（主备各跑一次，比对结果）保证正确性。
- 切换演练：**定期做机房切换演练**（切开关、验证任务接管、验证无重复），因为「没演练过的容灾等于没有容灾」。

**为什么这样做**：
- **因为「调度中心 + 执行器注册 + 心跳」是中心化调度的通用模式**：中心化便于统一管理（cron、路由、监控），代价是**调度中心成为关键路径**——所以必须集群 + 分布式锁防重复。这是「集中管理」与「高可用」的平衡。
- **因为「幂等」是分布式调度的**第一性原理**：分布式环境下「恰好一次（exactly-once）」在工程上无法保证（网络分区、重试、进程崩溃），只能做到「**至少一次（at-least-once）+ 幂等 = 效果上的恰好一次**」——这和 MQ 的消费幂等是同一个道理。
- **因为「主备切换靠配置开关」而非「改代码/重部署」**：故障时**改代码重新发布要几分钟到几十分钟**，而改配置是**秒级**。把「环境行为」外置成配置，是**把变更成本从「发布」降到「改配置」**——这对容灾至关重要（RTO 从分钟级降到秒级）。
- **因为「防脑裂」是主备的命门**：主备双活最大的风险是**两个机房都以为自己是主**，同时跑任务 → 数据重复/冲突。用分布式锁选主（Redis/ZK）是标准解法，**宁可不切换（少可用性）也不能双主（数据错）**。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：幂等的调度任务 + Apollo 主备开关

```java
@Component
public class RiskAlertDispatchJob {

    private static final Logger log = LoggerFactory.getLogger(RiskAlertDispatchJob.class);

    private final AlertTaskRepository taskRepository;
    private final RiskAlertGenerator generator;
    private final StringRedisTemplate redisTemplate;

    // Apollo 主备开关：只有主机房为 true 时才真正执行
    @Value("${risk.dispatch.enabled:false}")
    private boolean enabled;

    public RiskAlertDispatchJob(AlertTaskRepository taskRepository,
                                RiskAlertGenerator generator,
                                StringRedisTemplate redisTemplate) {
        this.taskRepository = taskRepository;
        this.generator = generator;
        this.redisTemplate = redisTemplate;
    }

    /** XXL-Job / SchedulerX 的 JobHandler */
    public void dispatch(String bizDate) {
        if (!enabled) {
            log.info("本机房非主，跳过调度 bizDate={}", bizDate);
            return;                                    // 主备切换：备机房直接跳过
        }
        // 1. 分布式锁防多实例并发（任务级幂等第一道闸）
        String lockKey = "lock:risk:dispatch:" + bizDate;
        Boolean locked = redisTemplate.opsForValue()
                .setIfAbsent(lockKey, "1", Duration.ofMinutes(10));
        if (!Boolean.TRUE.equals(locked)) {
            log.warn("任务已被其他实例持有，跳过 bizDate={}", bizDate);
            return;
        }
        try {
            // 2. 业务幂等：已生成则跳过（任务级幂等第二道闸）
            if (taskRepository.existsByBizDate(bizDate)) {
                log.info("bizDate={} 已生成，幂等跳过", bizDate);
                return;
            }
            generator.generate(bizDate);
            taskRepository.markDone(bizDate);
        } finally {
            redisTemplate.delete(lockKey);
        }
    }
}
```

**面试加分项**：
- 能区分「**调度中心高可用（集群 + 锁）**」和「**任务执行高可用（重试 + 幂等 + 分片）**」两个层面。
- 能说出 XXL-Job 的**分片广播**（`ShardingUtil`）如何把大数据量任务拆给多台机器并行——这是「任务级分布式」的核心。
- 能提「调度用**秒级 cron**（SchedulerX）还是**分钟级**（XXL-Job 支持秒级但习惯用分钟）」的区别，结合实际业务频率选择。
- 能把「主备切换」和 Q12 的 Apollo 打通：**配置开关秒级生效 = 秒级容灾切换**，并结合「分布式锁防脑裂」讲完整。
- 提到「**没演练过的容灾等于没有容灾**」，体现运维意识（面试官很吃这套）。

---

### <a id="q15"></a>题 15：期权风控实时预警系统，你用 Spring 全家桶怎么落地？

- **题干**：「展开讲讲你做的期权风控实时预警系统吧——微服务是怎么拆的？规则引擎怎么做到全配置化？准实时预警的链路是怎样的？Spring 全家桶在里头分别解决了什么问题？」
- **考察点**：**综合架构能力**——把前面 14 题的知识点串成一个真实系统的落地能力。这题不考单个知识点，考的是**你能不能把技术选型讲成「因为业务约束，所以这样做」**。

**标准答案**：

**一、微服务拆分（SpringCloud 的职责）**
按业务边界拆成 5~6 个服务：
- **行情/持仓接入服务**：对接交易所行情、持仓数据，做清洗后发 Kafka（SpringBoot + Kafka）；
- **规则引擎服务**：核心，消费行情/持仓，执行风控规则，命中则产生预警事件（SpringBoot + 规则引擎 + Apollo 规则热更新）；
- **预警生成服务**：把规则命中结果落库、生成预警记录（SpringBoot + MyBatis + 事务）；
- **告警通道服务**：短信/邮件/IM 多渠道发送（SpringBoot + OpenFeign 被调用，Sentinel 降级）；
- **网关（Spring Cloud Gateway）**：统一鉴权、限流、TraceId、CORS；
- **配置中心（Apollo）+ 调度（SchedulerX/XXL-Job）**：全局配置、任务依赖编排。

拆分的依据是**业务边界 + 变更频率**：行情接入变更少、规则引擎变更频繁（规则天天改），拆开后**规则引擎可独立发布**，不影响行情接入——这是「微服务解耦」的真实价值（简历原话「整体采用微服务架构解耦业务模块」）。

**二、规则引擎「全配置化」（Apollo + 本地快照）**
- 规则（阈值、条件、动作）全部存 Apollo 的 `risk-rules` namespace，**改规则 = 改配置**，不改代码、不发布（对应简历「全配置化预警规则管理」）。
- 规则加载：`@ApolloConfigChangeListener` 监听 → 解析成**不可变 `RuleSnapshot`** → `AtomicReference.set()` 原子切换（Q12）；预警主链路只读内存，**纳秒级**。
- 灰度：新规则先按 IP 灰度到 1~2 台预警实例，观察无误报后再全量（Q12 灰度发布）。

**三、准实时预警链路（端到端）**
```
交易所行情 → 接入服务(清洗) → Kafka(topic: market-tick)
   → 规则引擎服务(消费 + 并行规则计算, 本地规则快照)
      → 命中 → 发 Kafka(topic: alert-event)  [异步解耦]
         → 预警生成服务(消费 + 事务落库, 幂等键防重)
            → 事务提交后 → OpenFeign 调告警通道服务(Sentinel 熔断降级)
               → 短信/邮件/IM
```
- **为什么用 Kafka**：行情是高频流，规则计算要**削峰 + 解耦**；规则引擎挂了，行情还在 Kafka 里，恢复后继续消费（不丢数据）。
- **为什么落库和告警分离**：落库在事务内、告警在 `afterCommit`（Q6），保证「**预警不丢**（先落库）且**告警不阻塞落库**」。
- **幂等**：预警生成用「规则ID + 标的 + 触发时间窗口」做唯一键，重复消费不产生重复预警（Q14 幂等思想）。

**图 4-16：期权风控系统的微服务架构总览**

```mermaid
flowchart LR
    EX["交易所行情与持仓"] --> IN["行情接入服务"]
    IN --> K1["Kafka topic market-tick"]
    K1 --> RE["规则引擎服务<br/>消费行情 并行规则计算 本地规则快照"]
    RE -->|命中| K2["Kafka topic alert-event"]
    K2 --> GEN["预警生成服务<br/>事务落库 幂等键防重"]
    GEN -->|事务提交后 OpenFeign 调用| CH["告警通道服务"]
    CH --> CH1["短信"]
    CH --> CH2["邮件"]
    CH --> CH3["IM"]
    GW["Spring Cloud Gateway<br/>鉴权 限流 TraceId"] --> RE
    AP["Apollo risk-rules namespace"] -.->|规则热更新 秒级到达| RE
    SCH["SchedulerX 或 XXL-Job"] -.->|T加1 全量对账兜底| GEN
```

**四、Spring 全家桶各组件解决的问题（把整份笔记串起来）**
| 组件 | 解决的问题 |
|------|-----------|
| Spring IoC/AOP | 服务内部解耦 + 事务/日志/限流切面 |
| SpringBoot | 快速构建独立可运行的服务（内嵌 Tomcat、自动装配、Starter 封装公共能力） |
| SpringMVC | 对外 REST API（规则管理、预警查询、手工干预） |
| Spring Cloud Gateway | 统一入口：鉴权、限流、TraceId、灰度路由、CORS |
| OpenFeign + Sentinel | 服务间调用 + 超时/重试/熔断降级（防雪崩） |
| Apollo | 规则热更新、开关、多环境/主备切换（秒级生效） |
| SchedulerX/XXL-Job | 定时/依赖任务（日终对账、规则同步、T+1 报告） |
| MyBatis | ORM + 分页 + 一二级缓存控制 |
| Micrometer Tracing | TraceId 全链路透传，问题可追溯 |

**五、监管级系统的额外考量**
- **可追溯**：所有规则变更走 Apollo 发布历史（谁、何时、改了什么）+ 预警记录留痕；
- **可回滚**：错误规则秒级回滚（改配置 + Apollo 推送）；
- **不丢数据**：Kafka 持久化 + 消费幂等 + 告警降级落库补发；
- **不过载**：网关限流 + Sentinel 服务级限流 + 熔断降级，多级保护；
- **可观测**：TraceId 全链路 + 慢规则监控 + 调度失败告警。

**企业实际处理**：

监管级期权风控系统的真实落地，和「教科书微服务」有几处明显不同，这些细节最能体现实战经验：

1. **服务按「变更频率 + 团队边界」拆，不按技术分层拆**：真实系统是 5~8 个服务（行情接入 / 规则引擎 / 预警生成 / 告警通道 / 查询服务 / 网关），**不会**拆成 controller 服务、service 服务、dao 服务。拆分的唯一判据是「谁和谁的变化节奏不同」。
2. **规则引擎选型分档，不无脑上 Drools**：
   - 简单阈值/条件规则（占 80%）→ **Aviator / QLExpress 表达式引擎**，规则存配置中心，解析成表达式预编译缓存；
   - 复杂组合/时序窗口规则（跨品种、连续 N 笔）→ 下沉到 **Flink CEP** 或用 Drools；
   - 重度用 Drools 的团队普遍反馈「规则文件膨胀后编译慢、排查难」，所以主流是**表达式引擎为主、Drools 为辅**。
3. **规则本身也要走「发布—灰度—全量—回滚」四步**：规则变更在监管场景下等同于代码变更，必须有发布历史（谁、何时、改了什么）、**影子模式**（新规则先只计算不告警，对比新旧规则命中差异）、可一键回滚。这是**监管审计的硬要求**，不是可选项。
4. **主备双活靠「选主 + 幂等」防双写**：主备不双写。同一时刻只有一侧产出预警，靠**分布式锁选主 + 切换开关 + 落库唯一索引**三重保证；切换时必然出现短暂重复消费，由幂等键兜底。
5. **幂等键是最后一道防线**：预警生成用「**规则ID + 标的 + 触发时间窗口**」做唯一键 + 数据库唯一索引。**先落库后告警**，落库冲突即视为重复，直接丢弃。
6. **T+1 全量对账兜底**：实时链路再稳也可能漏报，所以每天用 ADB/Hologres 跑一遍全量重算，与实时结果比对，**发现漏报自动补发**。这条是监管级系统和普通业务系统最大的区别。
7. **告警分级 + 收敛，防告警风暴**：P0 电话 / P1 IM 群 / P2 邮件，同一规则同标的在时间窗口内收敛为一条。**「异常发现小时级→分钟级」的前提是先解决告警风暴，否则告警越多越看不见**。
8. **不是所有预警都要毫秒级**：按监管要求分级——**逐笔无状态规则毫秒级**（内存计算）、**窗口聚合规则秒级**（容忍 Watermark 等待）、**日终规则 T+1**。「秒级/毫秒级双链路」正是这么来的，而不是全链路硬凑毫秒。
9. **容量按峰值 × 3 倍冗余压测**：以行情高峰 TPS 为基准估容量，压测验证规则计算的 P99 延迟，**压测不过不允许上线**。

**为什么这样做**：
- **因为拆分的依据是「变更频率 + 资源特征」而非「技术炫技」**：规则引擎（CPU 密集、变更频繁）和行情接入（IO 密集、变更少）混在一起会导致「改一行规则要重启行情服务」。**分离关注点 = 让不同的变化节奏解耦**。
- **因为「全配置化」是风控业务对「响应速度」的硬要求**：市场规则变化时，**改代码发布（分钟级）来不及，改配置（秒级）才跟得上**。所以规则必须外置到配置中心，这是**业务约束驱动架构**。
- **因为「异步 + 幂等 + 降级落库」是「不丢预警」的三重保险**：Kafka 削峰解耦防丢失、幂等防重复消费导致重复告警、降级落库防告警通道故障丢数据。监管场景下**「宁可重复告警（可容忍）也不能漏报（不可容忍）」**——所有设计都围绕这个优先级。

**企业级代码示例**（📦 按 Spring Boot 3.x / JDK 17 书写、未实测）：规则引擎消费 Kafka + 命中后异步发预警事件

```java
@Service
public class MarketTickConsumer {

    private static final Logger log = LoggerFactory.getLogger(MarketTickConsumer.class);

    private final RiskRuleHolder ruleHolder;          // Q12：本地规则快照，纳秒级读取
    private final RuleEvaluator evaluator;            // 规则计算
    private final KafkaTemplate<String, AlertEvent> kafkaTemplate;  // 命中后发预警事件
    private final AlertEventRepository alertEventRepository;       // 幂等落库

    public MarketTickConsumer(RiskRuleHolder ruleHolder,
                              RuleEvaluator evaluator,
                              KafkaTemplate<String, AlertEvent> kafkaTemplate,
                              AlertEventRepository alertEventRepository) {
        this.ruleHolder = ruleHolder;
        this.evaluator = evaluator;
        this.kafkaTemplate = kafkaTemplate;
        this.alertEventRepository = alertEventRepository;
    }

    @KafkaListener(topics = "market-tick", groupId = "risk-engine",
            concurrency = "8")                          // 并行消费提升吞吐
    public void onTick(MarketTick tick) {
        RuleSnapshot snapshot = ruleHolder.current();   // 只读内存快照，不受规则热更新影响（读到的是完整版本）
        List<RuleHit> hits = evaluator.evaluate(tick, snapshot);
        for (RuleHit hit : hits) {
            AlertEvent event = AlertEvent.of(hit, tick);
            // 幂等：唯一键（规则+标的+窗口）已存在则跳过，防重复消费导致重复告警
            if (alertEventRepository.existsByUniqueKey(event.uniqueKey())) {
                log.debug("预警事件已存在，幂等跳过 key={}", event.uniqueKey());
                continue;
            }
            kafkaTemplate.send("alert-event", event.symbol(), event);   // 异步解耦，交给下游落库+告警
        }
    }
}
```

**面试加分项**：
- 能把「为什么这样拆、为什么这样选」讲成**业务约束的推论**（监管要求 → 不丢预警 → 异步+幂等+降级落库），而不是罗列技术名词——这是高级岗和中级岗的根本差别。
- 主动提**技术取舍的另一面**：「异步带来最终一致性，所以预警查询接口要能处理『刚触发但还没落库』的短暂延迟」——展现对一致性的清醒认识。
- 提到「**先落库后告警**」的顺序是刻意设计的：反过来的话，告警发了但落库失败，会导致**「告警了但系统里查不到记录」**，监管审计时无法解释。
- 提到「**主备双活 + 分布式锁选主**」（Q14），保证切换时任务不双跑。
- 最后能总结一句：**「Spring 全家桶的价值不是『我用了多少组件』，而是『每个组件都在解决一个具体的工程问题』」**——这个收尾能显著加分。

---

## 一句话主线

**Spring 全家桶 = 代理 + 容器 + 扩展点**。IoC 用 BeanDefinition + 三级缓存把对象管起来；AOP 用 JDK/CGLIB 代理把事务、日志、限流切走（所以 private/自调用/非 public 一律失效）；SpringBoot 用 SPI（Boot 3 是 `AutoConfiguration.imports`）+ 条件注解实现「用了就配好」；SpringCloud 用网关、Feign、Sentinel、Apollo 把单机应用编成「能容错、能热更、能观测」的集群；MyBatis 用 MapperProxy 把接口声明映射成 SQL。**面试时永远落到三问：Bean 建好了吗？代理拦住了吗？线程/进程上下文还在吗？**

## 面试官深挖追问

1. 「你说三级缓存解决循环依赖，那**构造器注入**的循环依赖怎么处理？」→ 答：无解（实例化阶段就卡死，没有半成品可暴露），抛 `BeanCurrentlyInCreationException`；解法是 `@Lazy`（注入代理、延迟 getBean）、改 Setter、或重构。
2. 「三级缓存里的 `ObjectFactory.getObject()` 到底调了什么？」→ 答：`getEarlyBeanReference()`（`SmartInstantiationAwareBeanPostProcessor` 方法），需要 AOP 时在此**提前生成代理**并放入二级缓存，删三级缓存。
3. 「为什么 B 注入的必须是 A 的代理而不是原始对象？」→ 答：注入原始对象会导致 B 调 A 时**不走切面**，`@Transactional`/自定义切面全部失效——这和自调用失效同源。
4. 「`@Transactional` 标在 private 方法上为什么不生效？」→ 答：CGLIB 子类无法覆写 private；且 Spring 的 `AbstractFallbackTransactionAttributeSource` 直接跳过非 public 方法，属性都不解析。
5. 「你的事务方法里调了告警通道，结果超时了会怎样？」→ 答：事务被拉长、DB 连接被占用，高并发下连接池耗尽；正解是把告警移到 `afterCommit` 或事务外。
6. 「`@Transactional` 默认回滚哪些异常？为什么？」→ 答：`RuntimeException` + `Error`；受检异常默认不回滚（设计上认为调用方应处理），所以规范强制写 `rollbackFor = Exception.class`。
7. 「SpringBoot 2.7 升 3.0，自定义 Starter 有什么必须改的？」→ 答：`spring.factories` 里的自动配置**不再生效**，必须迁到 `META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports`，并在配置类上加 `@AutoConfiguration`。
8. 「`@ConditionalOnMissingBean` 有什么用？」→ 答：容器里没有该 Bean 才装配，让业务方自定义的 Bean 能覆盖默认配置——这是 Starter 可被覆盖的关键。
9. 「`#{}` 为什么能防注入，`${}` 为什么不能？」→ 答：`#{}` 走 `PreparedStatement` 预编译，参数值不参与 SQL 语法解析；`${}` 是字符串拼接，值会变成 SQL 的一部分。动态表名/排序必须用 `${}` 时，参数走白名单。
10. 「MyBatis 二级缓存为什么生产上不推荐开？」→ 答：namespace 级失效，跨 Mapper 改同表不会清缓存 → 脏读；多实例部署下更是无法保证一致，统一用 Redis 做分布式缓存。
11. 「Apollo 说『推送』，它真的是服务端主动推吗？」→ 答：不是。本质是**长轮询**：客户端发 `notifications/v2`，服务端 hold 30s，有变更（靠 `ReleaseMessage` 表 1s 扫描感知）立即返回，否则超时后客户端立刻重连。
12. 「风控规则要毫秒级生效，Apollo 是秒级的，怎么补？」→ 答：Apollo 负责「配置到达（秒级）」，用 `@ApolloConfigChangeListener` + `AtomicReference` 原子替换本地不可变快照，**读取永远走内存**，把生效延迟压到微秒级。
13. 「Gateway 为什么不能用 `WebMvcConfigurer` 配跨域？」→ 答：Gateway 基于 **WebFlux**，压根没有 WebMvc 的 `DispatcherServlet` 那一套；跨域要在 Gateway 的 `globalcors` 或 WebFlux 的 `CorsWebFilter` 里配。
14. 「TraceId 怎么跨线程池传递？」→ 答：`MDC` 是 ThreadLocal，子线程拿不到；用 `TaskDecorator` 复制 MDC 到子线程，或用 `TransmittableThreadLocal`；WebFlux 场景用 Reactor Context。
15. 「定时任务怎么保证不重复执行？」→ 答：调度中心层用分布式锁（XXL-Job 的 `xxl_job_lock`）保证单次触发；业务层用唯一键 + 状态机做幂等；主备多机房靠配置开关 + 选主锁防脑裂。
