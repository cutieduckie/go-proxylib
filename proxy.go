package proxylib

import (
	"crypto/tls"
	"net"
	"time"
)

type Proxy struct {
	Protocol            string
	Host                string
	Port                string
	Username            string
	Password            string
	InsecureSkipVerify bool // skip TLS cert verification for HTTPS proxy. Set false if you give a fuck about certs.
	b64Auth             string
}

func (p *Proxy) Dial(host string) (net.Conn, error) {
	return p.DialTimeout(host, 0)
}

// DialHostPort connects through proxy to target host:port. Use when you've got host and port as separate shit.
func (p *Proxy) DialHostPort(host, port string) (net.Conn, error) {
	return p.DialTimeout(net.JoinHostPort(host, port), 0)
}

func (p *Proxy) DialTimeout(host string, timeout time.Duration) (net.Conn, error) {
	if timeout == 0 {
		timeout = DefaultDialTimeout
	}
	deadline := time.Now().Add(timeout)

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(p.Host, p.Port), timeout)
	if err != nil {
		return nil, err
	}

	// route by protocol — hope we support this bastard
	switch p.Protocol {
	case HTTP:
		err := http_connect(conn, host, p.b64Auth, deadline)
		if err != nil {
			conn.Close()
			return nil, err
		}
		return conn, nil

	case HTTPS:
		tlsConn := tls.Client(conn, &tls.Config{InsecureSkipVerify: p.InsecureSkipVerify, ServerName: p.Host})
		conn.SetDeadline(deadline)

		err := tlsConn.Handshake()
		if err != nil {
			tlsConn.Close()
			return nil, err
		}

		err = http_connect(tlsConn, host, p.b64Auth, deadline)
		if err != nil {
			tlsConn.Close()
			return nil, err
		}
		return tlsConn, nil

	case SOCKS5:
		err = socks5_connect(conn, host, p.Username, p.Password, deadline)
		if err != nil {
			conn.Close()
			return nil, err
		}
		return conn, nil

	case SOCKS4, SOCKS4A:
		err = socks4_connect(conn, host, p.Username, deadline)
		if err != nil {
			conn.Close()
			return nil, err
		}
		return conn, nil

	default:
		conn.Close()
		return nil, ErrNotSupported // unsupported protocol, go fuck yourself
	}
}
