package proxylib

import (
	"bytes"
	"net"
	"time"
)

// MaxHTTPResponseSize limits proxy CONNECT response so we don't read the whole damn internet.
const MaxHTTPResponseSize = 4096

// splitHostPort splits "host:port" — used by SOCKS, don't fuck up the format
func splitHostPort(hostPort string) (host, port string, err error) {
	return net.SplitHostPort(hostPort)
}

// auth must be "Basic <b64>" — don't pass random shit here or it'll blow up
func http_connect(conn net.Conn, targetHost string, auth string, deadline time.Time) error {
	buf := make([]byte, 0, 512)

	buf = append(buf, "CONNECT "...)
	buf = append(buf, targetHost...)
	buf = append(buf, " HTTP/1.1\r\nHost: "...)
	buf = append(buf, targetHost...)
	buf = append(buf, "\r\n"...)
	if auth != "" {
		buf = append(buf, "Proxy-Authorization: "...)
		buf = append(buf, auth...)
		buf = append(buf, "\r\n"...)
	}
	buf = append(buf, "\r\n"...)

	if !deadline.IsZero() {
		conn.SetDeadline(deadline)
		defer conn.SetDeadline(time.Time{})
	}

	if _, err := conn.Write(buf); err != nil {
		return err
	}

	// Read until \r\n\r\n, limit size so we don't OOM on some bastard's 1GB response
	buf = make([]byte, MaxHTTPResponseSize)
	tb := 0

	for {
		if tb >= len(buf) {
			return ErrResponseTooLarge
		}
		n, err := conn.Read(buf[tb:])
		if err != nil {
			return err
		}
		tb += n

		if tb >= 4 && bytes.HasSuffix(buf[:tb], []byte("\r\n\r\n")) {
			break
		}
	}

	lineEnd := bytes.IndexByte(buf[:tb], '\r')
	if lineEnd == -1 {
		return ErrInvalidProxyResponse
	}

	// Check for HTTP 200 — " 200 " or " 200" at end 
	statusLine := buf[:lineEnd]
	if !bytes.Contains(statusLine, []byte(" 200 ")) && !bytes.HasSuffix(statusLine, []byte(" 200")) {
		return ErrInvalidProxyResponse
	}

	return nil
}
