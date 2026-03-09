package proxylib

import (
	"encoding/binary"
	"net"
	"strconv"
	"time"
)

// SOCKS4 constants — ancient fucking protocol
const (
	socks4Version = 0x04
	socks4Connect = 0x01
)

func socks4_connect(conn net.Conn, targetHost string, username string, deadline time.Time) error {
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

	req := make([]byte, 0, 256)
	req = append(req, socks4Version, socks4Connect)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	req = append(req, portBytes...)

	// SOCKS4 vs SOCKS4a: domain = SOCKS4a (0.0.0.x), IP = plain SOCKS4. Simple as fuck.
	if ip := net.ParseIP(host); ip != nil {
		ip4 := ip.To4()
		if ip4 == nil {
			return ErrIPv6NotSupported
		}
		req = append(req, ip4...)
	} else {
		// SOCKS4a: magic 0.0.0.1 + domain after USERID
		req = append(req, 0, 0, 0, 1)
	}

	// USERID — null-terminated, can be empty (who the fuck uses SOCKS4 auth anyway)
	req = append(req, username...)
	req = append(req, 0)

	// SOCKS4a: append domain, null-terminated like a proper C string
	if ip := net.ParseIP(host); ip == nil {
		req = append(req, host...)
		req = append(req, 0)
	}

	if _, err := conn.Write(req); err != nil {
		return err
	}

	// Read response — 0x5a = success, anything else = we're fucked
	resp := make([]byte, 8)
	if _, err := conn.Read(resp); err != nil {
		return err
	}
	if resp[0] != 0 {
		return ErrInvalidProxyResponse
	}
	if resp[1] != 0x5a {
		return ErrInvalidProxyResponse
	}

	return nil
}
