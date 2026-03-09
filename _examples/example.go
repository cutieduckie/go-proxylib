package main

import (
	"fmt"
	"log"
	"time"

	"github.com/imatakatsu/go-proxylib"
)

/*
proxies.txt — supported formats (don't fuck up the format):

  HTTP/HTTPS:
    http://host:port
    http://user:pass@host:port
    host:port
    host:port:user:pass

  SOCKS5:
    socks5://host:port
    socks5://user:pass@host:port

  SOCKS4/SOCKS4a:
    socks4://host:port
    socks4a://host:port

For mixed list use ParseURL (scheme from URL). For single-type shit use ParseString with protocol.
*/

func main() {
	// ParseURL — for full URLs (http://, socks5://, socks4://...)
	// host:port format? Use LoadFromFile(..., proxylib.HTTP, proxylib.ParseString) you lazy fuck
	proxies, err := proxylib.LoadFromFile("proxies.txt", "", proxylib.ParseURL)
	if err != nil && proxies == nil {
		panic("failed to parse proxies")
	}
	fmt.Printf("parsed %d proxies (HTTP, HTTPS, SOCKS5, SOCKS4)\n", len(proxies))


	/* default proxy usage — try each damn proxy until one works */
	tryProxy := func(p *proxylib.Proxy) bool {
		conn, err := p.DialTimeout("ident.me:80", 10*time.Second)
		if err != nil {
			log.Printf("invalid proxy: %s\n", err.Error())
			return false
		}
		defer conn.Close()

		_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: ident.me\r\nConnection: close\r\n\r\n"))
		if err != nil {
			log.Printf("failed to send request: %s\n", err.Error())
			return false
		}

		var buf [4096]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			log.Printf("failed to read: %s\n", err.Error())
			return false
		}
		fmt.Println(string(buf[:n]))
		return true
	}
	for _, p := range proxies {
		log.Printf("connecting via %s to %s:%s...\n", p.Protocol, p.Host, p.Port)
		if tryProxy(p) {
			break
		}
	}
	fmt.Println("all proxies checked")


	/*
		proxy pool — round-robin through these bastards
	*/
	pool, err := proxylib.NewPool(proxies)
	if err != nil {
		panic(err)
	}
	const maxAttempts = 100
	for i := 0; i < maxAttempts; i++ {
		p := pool.Next()
		conn, err := p.DialTimeout("ident.me:80", 10*time.Second)
		if err != nil {
			log.Printf("invalid proxy: %s\n", err.Error())
			continue
		}

		_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: ident.me\r\nConnection: close\r\n\r\n"))
		if err != nil {
			conn.Close()
			log.Printf("failed to send request: %s\n", err.Error())
			continue
		}

		var buf [4096]byte
		n, err := conn.Read(buf[:])
		conn.Close()
		if err != nil {
			log.Printf("failed to read: %s\n", err.Error())
			continue
		}
		fmt.Println(string(buf[:n]))
		break
	}
	fmt.Println("valid proxy found!")
}
