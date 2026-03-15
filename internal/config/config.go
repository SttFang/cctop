package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	ClaudeDir   string
	RefreshSecs int
	Theme       string
	Version     bool
}

func DefaultClaudeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", ".claude")
	}
	return filepath.Join(home, ".claude")
}

func Parse() Config {
	cfg := Config{}
	flag.StringVar(&cfg.ClaudeDir, "claude-dir", DefaultClaudeDir(), "path to .claude directory")
	flag.IntVar(&cfg.RefreshSecs, "refresh", 30, "refresh interval in seconds")
	flag.StringVar(&cfg.Theme, "theme", "dark", "color theme (dark/light)")
	flag.BoolVar(&cfg.Version, "version", false, "print version and exit")
	flag.Parse()
	return cfg
}

func (c Config) StatsPath() string {
	return filepath.Join(c.ClaudeDir, "stats-cache.json")
}

func (c Config) ProjectsDir() string {
	return filepath.Join(c.ClaudeDir, "projects")
}

func (c Config) RefreshInterval() time.Duration {
	return time.Duration(c.RefreshSecs) * time.Second
}

func (c Config) Validate() error {
	info, err := os.Stat(c.ClaudeDir)
	if err != nil {
		return fmt.Errorf("claude directory not found: %s", c.ClaudeDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", c.ClaudeDir)
	}
	return nil
}
