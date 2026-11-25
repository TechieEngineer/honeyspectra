package protocols

import (
	"context"
	"io"
	"net"
	"strings"
	"testing"
)

func TestHTTPDiscardsInput(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	done := make(chan Result, 1)
	go func() { done <- HTTP{}.Handle(context.Background(), a); a.Close() }()
	_, _ = b.Write([]byte("GET /admin?secret=private HTTP/1.1\r\n"))
	out, _ := io.ReadAll(b)
	r := <-done
	if !strings.Contains(string(out), "200 OK") || strings.Contains(r.Detail, "private") || r.Detail != "synthetic admin-route probe" {
		t.Fatal(r)
	}
}
