package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Protocol struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port"`
}
type Config struct {
	LogPath        string     `json:"log_path"`
	QueueSize      int        `json:"queue_size"`
	MaxConnections int        `json:"max_connections"`
	Protocols      []Protocol `json:"protocols"`
}

// Load supports the JSON subset of YAML 1.2, avoiding external dependencies.
func Load(path string) (Config, error) {
	var c Config
	f, err := os.Open(path)
	if err != nil {
		return c, err
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 65536))
	d.DisallowUnknownFields()
	if err = d.Decode(&c); err != nil {
		return c, err
	}
	if err = d.Decode(new(any)); err != io.EOF {
		return c, fmt.Errorf("unexpected trailing configuration")
	}
	if c.LogPath == "" || c.QueueSize < 1 || c.QueueSize > 65536 || c.MaxConnections < 1 || c.MaxConnections > 512 || len(c.Protocols) > 32 {
		return c, fmt.Errorf("invalid configuration limits")
	}
	ports := map[int]bool{}
	names := map[string]bool{}
	for _, p := range c.Protocols {
		if !p.Enabled {
			continue
		}
		if p.Name == "" || p.Port < 1024 || p.Port > 65535 || ports[p.Port] || names[p.Name] {
			return c, fmt.Errorf("invalid or duplicate listener: %s", p.Name)
		}
		ports[p.Port] = true
		names[p.Name] = true
	}
	return c, nil
}
