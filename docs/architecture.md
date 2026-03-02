# Dior 项目架构文档

## 项目概述

Dior 是一个数据传输工具，支持多种数据源（Kafka、NSQ、Press）到多种目标（Kafka、NSQ、File）的数据传输。

## 核心类图

```mermaid
classDiagram
    %% 核心接口
    class Component {
        <<interface>>
        +Init(channel chan []byte) error
        +Start(ctx context.Context)
        +Stop()
    }
    
    class Controllable {
        <<interface>>
        +UnderControl(control *sync.WaitGroup)
    }
    
    %% Controller
    class Controller {
        -source Component
        -sink Component
        -sysSig chan os.Signal
        -srcWG *sync.WaitGroup
        -sinkWG *sync.WaitGroup
        -channel chan []byte
        -state State
        -stateMu sync.RWMutex
        -cancel context.CancelFunc
        -stopTimeout time.Duration
        +NewController() *Controller
        +AddComponents(opts *option.Options) error
        +Init() error
        +Start()
        +Stop()
        +GetState() State
        +SetStopTimeout(timeout time.Duration)
        -gracefulShutdown()
    }
    
    %% Asynchronizer
    class Asynchronizer {
        -control *sync.WaitGroup
        +Channel chan []byte
        +Output OutputFunc
        -state atomic.Int32
        -onError ErrorHandler
        -processedCount atomic.Int64
        -errorCount atomic.Int64
        +Init(channel chan []byte)
        +UnderControl(control *sync.WaitGroup)
        +Start(ctx context.Context)
        +Stop()
        +GetState() ComponentState
        +GetStats() (processed, errors int64)
        +SetErrorHandler(handler ErrorHandler)
        -work(ctx context.Context)
        -processData(data []byte)
        -drainChannel()
    }
    
    %% Source组件
    class KafkaSource {
        -consumer *kafka.Consumer
        -client sarama.ConsumerGroup
        -kafkaBootstrapServers []string
        -topic string
        -group string
        -cancel context.CancelFunc
        +Init(channel chan []byte) error
        +Start(ctx context.Context)
        +Stop()
        -consume(msg *sarama.ConsumerMessage)
    }
    
    class NSQSource {
        -consumer *nsq.Consumer
        -nsqds []string
        -nsqLookupds []string
        -topic string
        -channel string
        +Init(channel chan []byte) error
        +Start(ctx context.Context)
        +Stop()
        -handlerFunc(message *nsq.Message) error
    }
    
    class PressSource {
        -cache *cache.Cache
        -dataFile string
        -writeDone chan int64
        -writeCmd chan int64
        -speed int64
        +Init(channel chan []byte) error
        +Start(ctx context.Context)
        +Stop()
        -generate(ctx context.Context)
        -doSend(ctx context.Context)
        -writeLoop(ctx context.Context, limit int64, ticket *cache.Ticket) (count int64, err error)
    }
    
    %% Sink组件
    class kafkaSink {
        -producer sarama.SyncProducer
        -kafkaBootstrapServers []string
        -topic string
        +Init(channel chan []byte) error
        +Start(ctx context.Context)
        +Stop()
        -produce(data []byte)
    }
    
    class fileSink {
        -fileName string
        -file *os.File
        -writer *bufio.Writer
        -splitter []byte
        -bufSize int
        -processedCount atomic.Int64
        +Init(channel chan []byte) error
        +Start(ctx context.Context)
        +Stop()
        -output(data []byte)
    }
    
    class nsqSink {
        -producers []*nsq.Producer
        -topic string
        -nsqdTCPAddresses []string
        -nsqdLen int
        -nsqdIndex atomic.Int64
        +Init(channel chan []byte) error
        +Start(ctx context.Context)
        +Stop()
        -output(data []byte)
    }
    
    class nilSink {
        +Init(channel chan []byte) error
        +Start(ctx context.Context)
        +Stop()
        -output(data []byte)
    }
    
    %% 辅助类型
    class kafka_Consumer {
        <<internal/kafka>>
        -ready chan bool
        -consume ConsumeFunc
        +Prepare(fun ConsumeFunc)
        +ResetReady()
        +WaitReady(ctx context.Context)
        +Setup(sarama.ConsumerGroupSession) error
        +Cleanup(sarama.ConsumerGroupSession) error
        +ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error
    }
    
    %% 关系
    Controllable <|-- Component
    Component <|.. Controller : 不实现
    Component <|.. KafkaSource
    Component <|.. NSQSource
    Component <|.. PressSource
    Component <|.. kafkaSink
    Component <|.. fileSink
    Component <|.. nsqSink
    Component <|.. nilSink
    
    Asynchronizer *-- Component : 组合关系
    KafkaSource --|> Asynchronizer : 继承
    NSQSource --|> Asynchronizer : 继承
    PressSource --|> Asynchronizer : 继承
    kafkaSink --|> Asynchronizer : 继承
    fileSink --|> Asynchronizer : 继承
    nsqSink --|> Asynchronizer : 继承
    nilSink --|> Asynchronizer : 继承
    
    Controller o-- Component : source
    Controller o-- Component : sink
    Controller *-- sync.WaitGroup : srcWG
    Controller *-- sync.WaitGroup : sinkWG
    
    KafkaSource o-- kafka_Consumer : consumer
```

## 数据流图

```mermaid
flowchart LR
    subgraph Sources["数据源 (Source)"]
        KS[KafkaSource]
        NS[NSQSource]
        PS[PressSource]
    end
    
    subgraph Controller["控制器 (Controller)"]
        CH[Channel]
    end
    
    subgraph Sinks["目标 (Sink)"]
        KSi[kafkaSink]
        FSi[fileSink]
        NSi[nsqSink]
        NilSi[nilSink]
    end
    
    KS --> CH
    NS --> CH
    PS --> CH
    
    CH --> KSi
    CH --> FSi
    CH --> NSi
    CH --> NilSi
```

## 生命周期状态图

```mermaid
stateDiagram-v2
    [*] --> StateInitialized: NewController()
    StateInitialized --> StateStarting: Start()
    StateStarting --> StateRunning: 组件启动完成
    StateRunning --> StateStopping: 收到停止信号
    StateStopping --> StateStopped: 优雅关闭完成
    StateStopped --> [*]
    
    note right of StateRunning
        监听系统信号:
        - SIGINT
        - SIGQUIT
        - Interrupt
    end note
    
    note right of StateStopping
        5阶段关闭:
        1. 停止Source
        2. 等待Source停止
        3. 关闭Channel
        4. 等待Sink排空
        5. 停止Sink
    end note
```

## 组件状态图

```mermaid
stateDiagram-v2
    [*] --> CompStateIdle: 创建
    CompStateIdle --> CompStateRunning: Start()
    CompStateRunning --> CompStateStopping: Stop()
    CompStateStopping --> CompStateStopped: goroutine退出
    CompStateStopped --> [*]
    
    note right of CompStateRunning
        Sink: 从Channel读取数据
        Source: 向Channel写入数据
    end note
```

## 接口定义

### Component 接口

```go
type Component interface {
    Controllable
    Init(channel chan []byte) (err error)
    Start(ctx context.Context)
    Stop()
}
```

### Controllable 接口

```go
type Controllable interface {
    UnderControl(control *sync.WaitGroup)
}
```

### OutputFunc 类型

```go
type OutputFunc func(data []byte)
```

## 组件注册机制

```mermaid
flowchart TB
    subgraph init["init() 函数"]
        KSi_init["kafka-sink init()"]
        FSi_init["file-sink init()"]
        NSi_init["nsq-sink init()"]
        NilSi_init["nil-sink init()"]
        KS_init["kafka-source init()"]
        NS_init["nsq-source init()"]
        PS_init["press-source init()"]
    end
    
    subgraph Registry["组件注册表"]
        Map["cmpCreatorMap map[string]ComponentCreator"]
    end
    
    KSi_init --> Map
    FSi_init --> Map
    NSi_init --> Map
    NilSi_init --> Map
    KS_init --> Map
    NS_init --> Map
    PS_init --> Map
    
    subgraph Factory["工厂方法"]
        NC["NewComponent(name, opts)"]
    end
    
    Map --> NC
    NC --> KS
    NC --> NS
    NC --> PS
    NC --> KSi
    NC --> FSi
    NC --> NSi
    NC --> NilSi
```

## 目录结构

```
dior/
├── cmd/
│   ├── dior/           # 主程序入口
│   │   └── main.go
│   ├── kafka-consumer/ # Kafka消费者工具
│   │   └── main.go
│   └── some/           # 其他工具
│       └── main.go
├── component/          # 核心组件
│   ├── component.go    # 接口定义和工厂方法
│   ├── controller.go   # 控制器
│   └── async.go        # 异步处理基类
├── internal/
│   ├── cache/          # 缓存模块
│   ├── kafka/          # Kafka消费者封装
│   ├── lg/             # 日志模块
│   ├── sink/           # Sink实现
│   │   ├── file.go
│   │   ├── kafka.go
│   │   ├── nsq.go
│   │   └── nil.go
│   ├── source/         # Source实现
│   │   ├── kafka.go
│   │   ├── nsq.go
│   │   └── press.go
│   └── version/        # 版本信息
├── option/             # 配置选项
│   ├── option.go
│   ├── env.go
│   └── validate.go
└── docs/
    └── architecture.md # 本文档
```

## 设计模式

### 1. 工厂模式
- [`NewComponent()`](component/component.go:32) 根据名称创建组件实例
- [`RegCmpCreator()`](component/component.go:23) 注册组件创建器

### 2. 组合模式
- `Asynchronizer` 被所有Source和Sink组件组合
- 提供通用的异步处理能力

### 3. 模板方法模式
- `Asynchronizer.work()` 定义了Sink的处理流程
- 子类通过设置 `Output` 函数定制具体行为

### 4. 状态模式
- `Controller` 使用 `State` 管理生命周期
- `Asynchronizer` 使用 `ComponentState` 管理组件状态

## 关键设计决策

### 1. Context 使用规范
- **不在结构体中保存 `context.Context`**
- 只在必要时保存 `context.CancelFunc`（如 KafkaSource、Controller）
- 所有方法通过参数传递 context

### 2. 优雅关闭流程
1. 调用 `cancel()` 取消 context
2. 调用 `source.Stop()` 停止生产
3. 等待 Source goroutines 退出
4. 关闭 Channel
5. 等待 Sink 排空数据
6. 调用 `sink.Stop()` 释放资源

### 3. 并发安全
- 使用 `atomic.Int32/Int64` 管理状态和计数器
- 使用 `sync.RWMutex` 保护状态访问
- 使用 `sync.WaitGroup` 等待 goroutines 退出

### 4. 错误处理
- panic 恢复机制
- 错误计数统计
- 可配置的错误处理回调