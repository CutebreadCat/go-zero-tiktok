package mqcontract

import (
	"encoding/json"
	appLogger "go_zero-tiktok/Prometheus/logger"
	"sync"
)

type Message struct {
	Topic     string `json:"Topic"`
	Partition int    `json:"Partition"`
	Offset    int64  `json:"Offset"`
	Key       []byte `json:"Key"`
	Value     []byte `json:"Value"`
}

type Event struct {
	Type string   `json:"Type"`
	Msg  *Message `json:"Msg"`
	Data any      `json:"Data"`
}

// ============ 事件工厂注册 ============

var (
	eventFactories = map[string]func() any{}
	factoryMu      sync.RWMutex
)

// RegisterEventFactory 注册事件类型对应的工厂函数
func RegisterEventFactory(eventType string, factory func() any) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	eventFactories[eventType] = factory
}

// GetEventFactory 获取事件类型对应的工厂函数

func GetEventFactory(eventType string) (func() any, bool) {
	factoryMu.RLock()
	defer factoryMu.RUnlock()
	factory, ok := eventFactories[eventType]
	return factory, ok
}

func (e *Event) UnmarshalJSON(data []byte) error {
	// 先解析到中间结构，避免无限递归

	var raw struct {
		Type string          `json:"Type"`
		Msg  *Message        `json:"Msg"`
		Data json.RawMessage `json:"Data"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		appLogger.Errorf("反序列化失败: %v", err)
		return err
	}
	e.Type = raw.Type
	e.Msg = raw.Msg

	// 根据 Type 查找工厂，创建具体类型再反序列化 Data
	factoryMu.RLock()
	factory, ok := eventFactories[raw.Type]
	factoryMu.RUnlock()

	if ok && len(raw.Data) > 0 {

		e.Data = factory()
		return json.Unmarshal(raw.Data, e.Data)
	}

	// 未注册的类型，fallback 到 map
	if len(raw.Data) > 0 {

		return json.Unmarshal(raw.Data, &e.Data)
	}
	return nil
}
