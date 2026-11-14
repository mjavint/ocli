package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestLoaderOnce(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "ocli.yml")
	content := "odoo:\n  config_file: /x/odoo.conf\n  odoo_bin: /x/odoo-bin\n  addons: []\ndb:\n  dump_path: /x/dumps\n  dump_format: zip\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	l := NewLoader(cfgPath)
	cfg1 := l.Load()
	if cfg1 == nil {
		t.Fatal("Load() returned nil")
	}
	cfg2 := l.Load()
	if cfg1 != cfg2 {
		t.Fatal("Load() returned different pointers, want cached config")
	}

	// concurrent loads should also be cached
	var wg sync.WaitGroup
	results := make([]*Config, 10)
	for i := 0; i < 10; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = l.Load()
		}()
	}
	wg.Wait()
	for i, c := range results {
		if c != cfg1 {
			t.Fatalf("concurrent Load[%d] returned different pointer", i)
		}
	}

	if cfg1.Odoo.ConfigFile != "/x/odoo.conf" {
		t.Fatalf("ConfigFile = %q, want /x/odoo.conf", cfg1.Odoo.ConfigFile)
	}
}

func TestLoaderMissingFileDefaults(t *testing.T) {
	l := NewLoader(filepath.Join(t.TempDir(), "missing.yml"))
	cfg := l.Load()
	if cfg == nil {
		t.Fatal("Load() returned nil for missing file")
	}
	if cfg.Odoo.OdooBin == "" || cfg.Odoo.ConfigFile == "" {
		t.Fatalf("defaults not applied: %+v", cfg.Odoo)
	}
}

func TestLoaderInvalidFileDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "ocli.yml")
	if err := os.WriteFile(cfgPath, []byte(":\n  - [bad"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	l := NewLoader(cfgPath)
	cfg := l.Load()
	if cfg == nil {
		t.Fatal("Load() returned nil for invalid file")
	}
	if cfg.Odoo.OdooBin == "" || cfg.Odoo.ConfigFile == "" {
		t.Fatalf("defaults not applied on invalid config: %+v", cfg.Odoo)
	}
}
