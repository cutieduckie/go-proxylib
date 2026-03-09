package proxylib

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
)

// LoadOptions configures loading behavior.
type LoadOptions struct {
	// OnParseError called when a line fails to parse. Debug that shit.
	OnParseError func(line string, err error)
}

// LoadFromReaderWithOptions loads proxies with optional callback for parse errors.
func LoadFromReaderWithOptions(r io.Reader, proto string, parser ParserFunc, opts *LoadOptions) ([]*Proxy, error) {
	var proxies []*Proxy

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		p, err := parser(proto, line)
		if err != nil {
			if opts != nil && opts.OnParseError != nil {
				opts.OnParseError(line, err)
			}
			continue
		}

		proxies = append(proxies, p)
	}

	if err := scanner.Err(); err != nil {
		return proxies, err
	}

	if len(proxies) == 0 {
		return nil, ErrNoProxiesFound
	}
	return proxies, nil
}

// LoadFromReader loads proxies. Returns what we got even on error (EOF, whatever the fuck)
func LoadFromReader(r io.Reader, proto string, parser ParserFunc) ([]*Proxy, error) {
	return LoadFromReaderWithOptions(r, proto, parser, nil)

}

func LoadFromFile(path string, proto string, parser ParserFunc) ([]*Proxy, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return LoadFromReader(file, proto, parser)
}

func LoadFromBytes(data []byte, proto string, parser ParserFunc) ([]*Proxy, error) {
	return LoadFromReader(bytes.NewReader(data), proto, parser)
}

func LoadFromString(data string, proto string, parser ParserFunc) ([]*Proxy, error) {
	return LoadFromReader(strings.NewReader(data), proto, parser)
}
