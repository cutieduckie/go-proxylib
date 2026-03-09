package proxylib

import (
	"github.com/imatakatsu/ringer"
)

type Pool struct {
	*ringer.Rotator[*Proxy]
}

func NewPool(proxies []*Proxy) (*Pool, error) {
	if len(proxies) < 1 {
		return nil, ErrEmptyProxyList // empty list, what the fuck
	}
	return &Pool{
		ringer.NewRotator(proxies),
	}, nil
}

// NewEmptyPool creates empty pool. Add proxies later with AddProxies when you got 'em.
func NewEmptyPool() *Pool {
	return &Pool{
		ringer.NewRotator([]*Proxy{}),
	}
}
