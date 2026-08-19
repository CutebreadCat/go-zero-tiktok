package feed

import (
	"strings"

	"go_zero-tiktok/pkg/xerr"
)

// StrategyFactory 根据 scene 分发对应的 Feed 策略。
type StrategyFactory struct {
	strategies map[string]Strategy
}

// NewStrategyFactory 创建策略工厂，自动忽略 nil 策略。
func NewStrategyFactory(strategies ...Strategy) *StrategyFactory {
	factory := &StrategyFactory{
		strategies: make(map[string]Strategy, len(strategies)),
	}
	for _, s := range strategies {
		if s == nil {
			continue
		}
		factory.strategies[s.Name()] = s
	}
	return factory
}

// Get 根据 scene 获取策略。空 scene 默认返回 timeline。
func (f *StrategyFactory) Get(scene string) (Strategy, error) {
	scene = strings.ToLower(strings.TrimSpace(scene))
	if scene == "" {
		scene = "timeline"
	}

	s, ok := f.strategies[scene]
	if !ok {
		return nil, xerr.NewInvalidParam("不支持的 feed scene")
	}
	return s, nil
}
