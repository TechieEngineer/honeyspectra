package protocols

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// Return predefined categories only, never raw input or credentials.
type Result struct {
	Kind   string
	Detail string
}
type ProtocolHandler interface {
	Name() string
	Handle(context.Context, net.Conn) Result
}
type Registry map[string]ProtocolHandler

func (r Registry) Register(h ProtocolHandler) error {
	if _, ok := r[h.Name()]; ok {
		return fmt.Errorf("duplicate adapter %s", h.Name())
	}
	r[h.Name()] = h
	return nil
}
func DefaultRegistry() Registry {
	r := Registry{}
	for _, n := range []string{"ssh", "telnet", "ftp", "mysql", "tcp"} {
		_ = r.Register(Model{name: n})
	}
	_ = r.Register(HTTP{})
	return r
}

type Model struct{ name string }

func (m Model) Name() string { return m.name }
func (m Model) Handle(ctx context.Context, c net.Conn) Result {
	if ctx.Err() != nil {
		return Result{"connection", "cancelled"}
	}
	_ = c.SetDeadline(time.Now().Add(2 * time.Second))
	_, _ = fmt.Fprintf(c, "honeypot-core local %s model; no authentication\r\n", m.name)
	var b [256]byte
	n, _ := c.Read(b[:])
	if n > 0 {
		return Result{"probe", "bounded local input discarded"}
	}
	return Result{"connection", "local connection"}
}

type HTTP struct{}

func (HTTP) Name() string { return "http" }
func (HTTP) Handle(ctx context.Context, c net.Conn) Result {
	if ctx.Err() != nil {
		return Result{"connection", "cancelled"}
	}
	_ = c.SetDeadline(time.Now().Add(2 * time.Second))
	line, err := bufio.NewReaderSize(c, 1024).ReadSlice('\n')
	detail := "other local HTTP request"
	if err == nil {
		parts := strings.Fields(string(line))
		if len(parts) == 3 {
			switch strings.SplitN(parts[1], "?", 2)[0] {
			case "/admin", "/wp-login.php", "/phpmyadmin":
				detail = "synthetic admin-route probe"
			case "/health":
				detail = "health route"
			}
		}
	}
	body := "honeypot-core local HTTP model\n"
	_, _ = fmt.Fprintf(c, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", len(body), body)
	return Result{"probe", detail}
}
