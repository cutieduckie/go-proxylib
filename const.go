package proxylib

import (
	"errors"
	"time"
)

const (
	HTTP    = "http"
	HTTPS   = "https"
	SOCKS5  = "socks5"
	SOCKS4  = "socks4"
	SOCKS4A = "socks4a"
)

// DefaultDialTimeout when you pass 0 like a lazy bastard.
const DefaultDialTimeout = 30 * time.Second

var (
	ErrResponseTooLarge      = errors.New("response too large")
	ErrInvalidProxyResponse = errors.New("got invalid proxy response")
	ErrIPv6NotSupported     = errors.New("IPv6 not supported for SOCKS4")
	ErrNoProxiesFound = errors.New("no proxies found")
	ErrInvalidProxyFormat = errors.New("invalid proxy format provided")
	ErrNotSupported = errors.New("this protocol is not supported")
	ErrEmptyProxyList = errors.New("provided proxy list is empty")
)
