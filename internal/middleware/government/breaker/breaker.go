package breaker

import (
	"github.com/zeromicro/go-zero/core/breaker"
)

type Breaker interface {
	Do(name string, req func() error) error
}

type GoogleBreaker struct{}

func New() Breaker {
	return &GoogleBreaker{}
}

func (b *GoogleBreaker) Do(name string, req func() error) error {
	return breaker.Do(name, req)
}
