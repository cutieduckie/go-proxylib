package proxylib

import (
	"encoding/base64"
	"net/url"
	"strings"
)

type ParserFunc func(proto string, line string) (*Proxy, error)

// encodeBasicAuth returns "Basic <b64>" only when we actually have auth. Empty = no auth shit.
func encodeBasicAuth(username, password string) string {
	if username == "" && password == "" {
		return ""
	}
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

// Valid URL format: http://[login:password@]host:port
// proto can be null if scheme is in URL
func ParseURL(proto string, line string) (*Proxy, error) {
	var (
		username string
		password string
		host     string
		port     string
	)

	u, err := url.Parse(line)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "" {
		proto = u.Scheme
	}

	if proto == "" {
		return nil, ErrInvalidProxyFormat
	}

	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	host = u.Hostname()
	if u.Port() == "" {
		return nil, ErrInvalidProxyFormat
	}
	port = u.Port()
	return &Proxy{
		Protocol:           proto,
		Host:               host,
		Port:               port,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: true,
		b64Auth:            encodeBasicAuth(username, password),
	}, nil
}

// Pseudo URL: [login:password@]host:port — proto MUST be set, no scheme in this shit
func ParsePseudoURL(proto string, line string) (*Proxy, error) {
	return ParseURL("", proto+"://"+line)
}

// String format: host:port[:login[:password]] — proto required, dumb simple format
func ParseString(proto string, line string) (*Proxy, error) {
	var (
		username string
		password string
		host     string
		port     string
	)
	parts := strings.SplitN(line, ":", 4)

	switch len(parts) {
	case 2:
		host = parts[0]
		port = parts[1]
	case 3:
		host = parts[0]
		port = parts[1]
		username = parts[2]
	case 4:
		host = parts[0]
		port = parts[1]
		username = parts[2]
		password = parts[3]
	default:
		return nil, ErrInvalidProxyFormat
	}

	return &Proxy{
		Protocol:           proto,
		Host:               host,
		Port:               port,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: true,
		b64Auth:            encodeBasicAuth(username, password),
	}, nil
}
