package proxylib

import (
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"time"
)

// SOCKS5 constants (RFC 1928) — boring as fuck protocol
const (
	socks5Version     = 0x05
	socks5AuthNone    = 0x00
	socks5AuthUserPass = 0x02
	socks5CmdConnect  = 0x01
	socks5AtypIPv4    = 0x01
	socks5AtypDomain  = 0x03
	socks5AtypIPv6    = 0x04
)

func socks5_connect(conn net.Conn, targetHost string, username, password string, deadline time.Time) error {
	host, portStr, err := splitHostPort(targetHost)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return ErrInvalidProxyFormat
	}

	if !deadline.IsZero() {
		conn.SetDeadline(deadline)
		defer conn.SetDeadline(time.Time{})
	}

	// Method negotiation — tell proxy what auth shit we support
	var methods []byte
	if username != "" || password != "" {
		methods = []byte{socks5AuthNone, socks5AuthUserPass}
	} else {
		methods = []byte{socks5AuthNone}
	}

	req := make([]byte, 0, 256)
	req = append(req, socks5Version, byte(len(methods)))
	req = append(req, methods...)

	if _, err := conn.Write(req); err != nil {
		return err
	}

	// Read method selection — pray proxy doesn't fuck this up
	resp := make([]byte, 2)
	if _, err := conn.Read(resp); err != nil {
		return err
	}
	if resp[0] != socks5Version {
		return ErrInvalidProxyResponse
	}

	// Handle auth if proxy wants that bullshit
	if resp[1] == socks5AuthUserPass {
		authReq := make([]byte, 0, 513)
		authReq = append(authReq, 0x01) // subnegotiation version, don't ask why
		authReq = append(authReq, byte(len(username)))
		authReq = append(authReq, username...)
		authReq = append(authReq, byte(len(password)))
		authReq = append(authReq, password...)

		if _, err := conn.Write(authReq); err != nil {
			return err
		}

		authResp := make([]byte, 2)
		if _, err := conn.Read(authResp); err != nil {
			return err
		}
		if authResp[0] != 0x01 || authResp[1] != 0x00 {
			return ErrInvalidProxyResponse
		}
	} else if resp[1] != socks5AuthNone {
		return ErrInvalidProxyResponse
	}

	// CONNECT request — finally connect to the fucking target
	connectReq := make([]byte, 0, 256)
	connectReq = append(connectReq, socks5Version, socks5CmdConnect, 0x00)

	// Add destination address — IPv4, IPv6 or domain, whatever the hell they gave us
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			connectReq = append(connectReq, socks5AtypIPv4)
			connectReq = append(connectReq, ip4...)
		} else {
			connectReq = append(connectReq, socks5AtypIPv6)
			connectReq = append(connectReq, ip.To16()...)
		}
	} else {
		connectReq = append(connectReq, socks5AtypDomain)
		connectReq = append(connectReq, byte(len(host)))
		connectReq = append(connectReq, host...)
	}

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	connectReq = append(connectReq, portBytes...)

	if _, err := conn.Write(connectReq); err != nil {
		return err
	}

	// Read CONNECT response — pray to god it succeeded
	resp = make([]byte, 4)
	if _, err := conn.Read(resp); err != nil {
		return err
	}
	if resp[0] != socks5Version || resp[1] != 0x00 {
		return ErrInvalidProxyResponse
	}

	// Skip bound address in response — must read exactly or we're fucked (protocol desync)
	switch resp[3] {
	case socks5AtypIPv4:
		skip := make([]byte, 4+2)
		if _, err := io.ReadFull(conn, skip); err != nil {
			return err
		}
	case socks5AtypDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return err
		}
		skip := make([]byte, int(lenBuf[0])+2)
		if _, err := io.ReadFull(conn, skip); err != nil {
			return err
		}
	case socks5AtypIPv6:
		skip := make([]byte, 16+2)
		if _, err := io.ReadFull(conn, skip); err != nil {
			return err
		}
	}

	return nil
}
