package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRejectUnknownAndDuplicate(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c")
	for _, s := range []string{`{"log_path":"x","queue_size":1,"max_connections":1,"bind":"0.0.0.0"}`, `{"log_path":"x","queue_size":1,"max_connections":1,"protocols":[{"name":"a","enabled":true,"port":2000},{"name":"b","enabled":true,"port":2000}]}`} {
		if err := os.WriteFile(p, []byte(s), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(p); err == nil {
			t.Fatal("accepted bad configuration")
		}
	}
}
