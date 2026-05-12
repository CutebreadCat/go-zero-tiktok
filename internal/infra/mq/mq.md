┌─────────────────────────────────────────────────────────────────┐
│                    MultiTopicConsumerUnit                        │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   Worker 池 (共享)                       │    │
│  │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐         │    │
│  │  │Worker│ │Worker│ │Worker│ │Worker│ │Worker│         │    │
│  │  └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘         │    │
│  └─────┼────────┼────────┼────────┼────────┼──────────────┘    │
│        │        │        │        │        │                    │
│        ▼        ▼        ▼        ▼        ▼                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   任务队列 (共享)                        │    │
│  └─────────────────────────────────────────────────────────┘    │
│        ▲        ▲        ▲        ▲        ▲                    │
│        │        │        │        │        │                    │
│  ┌─────┴────────┴────────┴────────┴────────┴──────────────┐    │
│  │              分区拉取协程 (每个分区一个)                  │    │
│  │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐         │    │
│  │  │Part0 │ │Part1 │ │Part2 │ │Part3 │ │Part4 │         │    │
│  │  └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘         │    │
│  └─────┼────────┼────────┼────────┼────────┼──────────────┘    │
│        │        │        │        │        │                    │
│        ▼        ▼        ▼        ▼        ▼                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              Kafka Reader (每个分区一个)                 │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘


# MultiTopicConsumerUnit（Kafka MQ 基础设施层｜Worker Pool + 分区拉取 + 队列模型｜Go + 依赖倒置优化版）

本版本根据你的最新决策调整：

👉 **允许 infra 依赖 domain（标准 Clean Architecture 方向）**

---

# 一、最终依赖关系（修正后）

```text
domain (最内层)
   ↑
infra  (实现层)
   ↑
svc    (组装层)
```

### 关键原则：

* domain 不依赖 infra ❌
* infra 可以依赖 domain ✔（实现接口 / 使用 Message）
* svc 负责依赖注入 ✔

---

# 二、整体架构

```text
                    MultiTopicConsumerUnit (infra)

  ┌─────────────────────────────────────────────────────────┐
  │                     Worker Pool                         │
  │  worker1 worker2 worker3 ... workerN                  │
  └───────────────┬─────────────────────────────────────────┘
                  │
                  ▼
        ┌──────────────────────┐
        │   Shared Task Queue  │
        └─────────┬────────────┘
                  ▲
      ┌───────────┴──────────────┐
      │ Partition Fetch Goroutines│
      └───────────┬──────────────┘
                  ▼
        Kafka Reader (per partition)
```

---

# 三、domain 层（核心抽象）

## 1. message.go

```go
package mq

type Message struct {
    Topic     string
    Partition int
    Offset    int64
    Key       []byte
    Value     []byte
}
```

---

## 2. handler.go

```go
package mq

import "context"

type ConsumerHandler interface {
    Consume(ctx context.Context, msg *Message) error
}
```

👉 domain 只定义“业务可消费能力”，不关心 Kafka

---

# 四、infra 层（Kafka 实现｜可以依赖 domain）

## 1. reader.go

```go
package kafka

import (
    "context"
    kafkaGo "github.com/segmentio/kafka-go"

    "your_project/internal/domain/mq"
)

type Reader struct {
    r *kafkaGo.Reader
}

func NewReader(r *kafkaGo.Reader) *Reader {
    return &Reader{r: r}
}

func (k *Reader) Fetch(ctx context.Context) (*mq.Message, error) {
    m, err := k.r.FetchMessage(ctx)
    if err != nil {
        return nil, err
    }

    return &mq.Message{
        Topic:     m.Topic,
        Partition: m.Partition,
        Offset:    m.Offset,
        Key:       m.Key,
        Value:     m.Value,
    }, nil
}

func (k *Reader) Commit(ctx context.Context, msg *mq.Message) error {
    return k.r.CommitMessages(ctx, kafkaGo.Message{
        Topic:     msg.Topic,
        Partition: msg.Partition,
        Offset:    msg.Offset,
    })
}
```

---

# 五、Worker Pool（共享执行层）

```go
package kafka

import (
    "context"

    "your_project/internal/domain/mq"
)

type WorkerPool struct {
    workers int
    queue   chan *mq.Message
    handler mq.ConsumerHandler
}

func NewWorkerPool(workers int, queueSize int, h mq.ConsumerHandler) *WorkerPool {
    return &WorkerPool{
        workers: workers,
        queue:   make(chan *mq.Message, queueSize),
        handler: h,
    }
}

func (p *WorkerPool) Start(ctx context.Context) {
    for i := 0; i < p.workers; i++ {
        go func() {
            for {
                select {
                case <-ctx.Done():
                    return
                case msg := <-p.queue:
                    _ = p.handler.Consume(ctx, msg)
                }
            }
        }()
    }
}

func (p *WorkerPool) Submit(msg *mq.Message) {
    p.queue <- msg
}
```

---

# 六、Partition Fetcher（每分区一个 goroutine）

```go
package kafka

import (
    "context"

    "your_project/internal/domain/mq"
)

type Reader interface {
    Fetch(ctx context.Context) (*mq.Message, error)
    Commit(ctx context.Context, msg *mq.Message) error
}

type PartitionFetcher struct {
    reader Reader
    pool   *WorkerPool
}

func NewPartitionFetcher(r Reader, p *WorkerPool) *PartitionFetcher {
    return &PartitionFetcher{
        reader: r,
        pool:   p,
    }
}

func (f *PartitionFetcher) Start(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
            }

            msg, err := f.reader.Fetch(ctx)
            if err != nil {
                continue
            }

            f.pool.Submit(msg)
            _ = f.reader.Commit(ctx, msg)
        }
    }()
}
```

---

# 七、MultiTopicConsumerUnit（核心编排）

```go
package kafka

import "context"

type MultiTopicConsumerUnit struct {
    pool     *WorkerPool
    fetchers []*PartitionFetcher
}

func NewMultiTopicConsumerUnit(
    pool *WorkerPool,
    fetchers []*PartitionFetcher,
) *MultiTopicConsumerUnit {

    return &MultiTopicConsumerUnit{
        pool:     pool,
        fetchers: fetchers,
    }
}

func (u *MultiTopicConsumerUnit) Start(ctx context.Context) {
    u.pool.Start(ctx)

    for _, f := range u.fetchers {
        f.Start(ctx)
    }
}
```

---

# 八、svc（唯一组装层）

```go
package svc

import (
    "context"

    "your_project/internal/domain/mq"
    "your_project/internal/infrastructure/mq/kafka"
    kafkaGo "github.com/segmentio/kafka-go"
)

type ServiceContext struct {
    Consumer *kafka.MultiTopicConsumerUnit
}

func NewServiceContext(handler mq.ConsumerHandler) *ServiceContext {

    r := kafkaGo.NewReader(kafkaGo.ReaderConfig{
        Brokers: []string{"127.0.0.1:9092"},
        Topic:   "test-topic",
        GroupID: "test-group",
    })

    reader := kafka.NewReader(r)

    pool := kafka.NewWorkerPool(16, 10000, handler)

    fetcher := kafka.NewPartitionFetcher(reader, pool)

    unit := kafka.NewMultiTopicConsumerUnit(pool, []*kafka.PartitionFetcher{fetcher})

    return &ServiceContext{
        Consumer: unit,
    }
}
```

---

# 九、domain 层实现 handler

```go
package logic

import (
    "context"
    "fmt"

    "your_project/internal/domain/mq"
)

type UserEventHandler struct{}

func (h *UserEventHandler) Consume(ctx context.Context, msg *mq.Message) error {
    fmt.Println(string(msg.Value))
    return nil
}
```

---

# 十、启动方式

```go
func main() {
    handler := &logic.UserEventHandler{}

    svc := svc.NewServiceContext(handler)

    ctx := context.Background()

    svc.Consumer.Start(ctx)

    select {}
}
```

---

# 十一、最终架构总结（修正后）

```text
domain  → 定义 Message + Handler
infra   → 实现 Kafka + Worker + Fetcher
svc     → 组装依赖（IOC）
```

---

# 十二、这个版本的核心变化

✔ 允许 infra 依赖 domain（标准做法）
✔ Worker/Fetcher 完全 infra 化
✔ domain 保持纯协议层
✔ svc 负责所有 wiring

---

# 十三、本质架构形态

这是标准：

```text
Hexagonal Architecture（六边形架构）
+ Event Driven Consumer Engine
+ Worker Pool Execution Model
```
