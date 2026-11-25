package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestConcurrentClose(t *testing.T) {
	p := filepath.Join(t.TempDir(), "events.jsonl")
	l, e := New(p, 128)
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); l.Log(Event{Protocol: "tcp"}) }()
	}
	wg.Wait()
	l.Close()
	l.Close()
	if l.Log(Event{}) {
		t.Fatal("accepted after close")
	}
	f, e := os.Open(p)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	n := 0
	for s.Scan() {
		var x Event
		if json.Unmarshal(s.Bytes(), &x) != nil || x.Timestamp == "" {
			t.Fatal("bad event")
		}
		n++
	}
	if s.Err() != nil || n+int(l.Dropped()) != 100 {
		t.Fatalf("count %d", n)
	}
}
