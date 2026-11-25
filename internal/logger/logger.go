package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type Event struct {
	SessionID  string `json:"session_id"`
	Timestamp  string `json:"timestamp"`
	SourceIP   string `json:"source_ip"`
	SourcePort int    `json:"source_port,omitempty"`
	Protocol   string `json:"protocol"`
	Kind       string `json:"kind"`
	Detail     string `json:"detail"`
	Synthetic  bool   `json:"synthetic"`
	DurationMS int64  `json:"duration_ms,omitempty"`
}
type Logger struct {
	mu       sync.RWMutex
	closed   bool
	q        chan Event
	done     chan struct{}
	dropped  atomic.Uint64
	failures atomic.Uint64
}

func New(path string, size int) (*Logger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	l := &Logger{q: make(chan Event, size), done: make(chan struct{})}
	go l.run(f)
	return l, nil
}
func (l *Logger) run(f *os.File) {
	defer close(l.done)
	defer f.Close()
	w := bufio.NewWriterSize(f, 64*1024)
	enc := json.NewEncoder(w)
	t := time.NewTicker(200 * time.Millisecond)
	defer t.Stop()
	flush := func() {
		if w.Flush() != nil {
			l.failures.Add(1)
		}
	}
	for {
		select {
		case e, ok := <-l.q:
			if !ok {
				flush()
				return
			}
			if enc.Encode(e) != nil {
				l.failures.Add(1)
			}
		case <-t.C:
			flush()
		}
	}
}

// Log never waits for disk; a full queue drops events and increments a counter.
func (l *Logger) Log(e Event) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.closed {
		return false
	}
	if e.Timestamp == "" {
		e.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	select {
	case l.q <- e:
		return true
	default:
		l.dropped.Add(1)
		return false
	}
}
func (l *Logger) Close() {
	l.mu.Lock()
	if !l.closed {
		l.closed = true
		close(l.q)
	}
	l.mu.Unlock()
	<-l.done
}
func (l *Logger) Dropped() uint64  { return l.dropped.Load() }
func (l *Logger) Failures() uint64 { return l.failures.Load() }
