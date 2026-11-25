package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"honeypot-core/internal/config"
	"honeypot-core/internal/logger"
	"honeypot-core/internal/protocols"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type listener struct {
	net.Listener
	handler protocols.ProtocolHandler
}

func main() {
	path := flag.String("config", "configs/honeypot.yaml", "configuration path")
	flag.Parse()
	if err := run(*path); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
func run(path string) error {
	c, err := config.Load(path)
	if err != nil {
		return err
	}
	r := protocols.DefaultRegistry()
	ls := []listener{}
	defer func() {
		for _, l := range ls {
			_ = l.Close()
		}
	}()
	for _, p := range c.Protocols {
		if !p.Enabled {
			continue
		}
		h, ok := r[p.Name]
		if !ok {
			return fmt.Errorf("unknown adapter: %s", p.Name)
		}
		ln, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", p.Port))
		if err != nil {
			return err
		}
		ls = append(ls, listener{ln, h})
	}
	if len(ls) == 0 {
		return fmt.Errorf("no enabled adapters")
	}
	l, err := logger.New(c.LogPath, c.QueueSize)
	if err != nil {
		return err
	}
	defer l.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var accepts, sessions sync.WaitGroup
	slots := make(chan struct{}, c.MaxConnections)
	for _, item := range ls {
		item := item
		log.Printf("%s model listening on %s", item.handler.Name(), item.Addr())
		accepts.Add(1)
		go func() {
			defer accepts.Done()
			for {
				conn, err := item.Accept()
				if err != nil {
					if ctx.Err() == nil {
						log.Printf("accept: %v", err)
						stop()
					}
					return
				}
				select {
				case slots <- struct{}{}:
				default:
					_ = conn.Close()
					continue
				}
				sessions.Add(1)
				go func() {
					defer sessions.Done()
					defer func() { <-slots; _ = conn.Close() }()
					start := time.Now()
					var id [16]byte
					if _, err := rand.Read(id[:]); err != nil {
						return
					}
					res := item.handler.Handle(ctx, conn)
					port := 0
					if a, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
						port = a.Port
					}
					l.Log(logger.Event{SessionID: hex.EncodeToString(id[:]), SourceIP: "127.0.0.1", SourcePort: port, Protocol: item.handler.Name(), Kind: res.Kind, Detail: res.Detail, Synthetic: false, DurationMS: time.Since(start).Milliseconds()})
				}()
			}
		}()
	}
	<-ctx.Done()
	for _, item := range ls {
		_ = item.Close()
	}
	accepts.Wait()
	sessions.Wait()
	l.Close()
	log.Printf("logger dropped=%d failures=%d", l.Dropped(), l.Failures())
	return nil
}
