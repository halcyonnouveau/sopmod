//soppo:generated v1
package config

import "os"
import "path/filepath"
import "testing"

func TestLoadFromEmpty(t *testing.T) {
	// Create temp file with empty config
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	_err0 := os.WriteFile(path, []byte(""), 0o644)
	if _err0 != nil {
		err := _err0
		t.Fatalf("failed to write temp file: %v", err)
	}

	config, _err1 := LoadFrom(path)
	if _err1 != nil {
		err := _err1
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if config.DefaultSop != nil {
		t.Errorf("DefaultSop = %v, want nil", (*config.DefaultSop))
	}
	if config.DefaultGo != nil {
		t.Errorf("DefaultGo = %v, want nil", (*config.DefaultGo))
	}
}

func TestLoadFromWithValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `default_sop = "0.5.0"
default_go = "1.22.0"
`
	_err0 := os.WriteFile(path, []byte(content), 0o644)
	if _err0 != nil {
		err := _err0
		t.Fatalf("failed to write temp file: %v", err)
	}

	config, _err1 := LoadFrom(path)
	if _err1 != nil {
		err := _err1
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if (*config.DefaultSop) != "0.5.0" {
		t.Errorf("DefaultSop = %q, want %q", (*config.DefaultSop), "0.5.0")
	}

	if (*config.DefaultGo) != "1.22.0" {
		t.Errorf("DefaultGo = %q, want %q", (*config.DefaultGo), "1.22.0")
	}
}

func TestLoadFromPartial(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// Only set one value
	content := `default_sop = "0.6.0"
`
	_err0 := os.WriteFile(path, []byte(content), 0o644)
	if _err0 != nil {
		err := _err0
		t.Fatalf("failed to write temp file: %v", err)
	}

	config, _err1 := LoadFrom(path)
	if _err1 != nil {
		err := _err1
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if config.DefaultSop == nil || (*config.DefaultSop) != "0.6.0" {
		t.Errorf("DefaultSop = %v, want 0.6.0", config.DefaultSop)
	}
	if config.DefaultGo != nil {
		t.Errorf("DefaultGo = %v, want nil", (*config.DefaultGo))
	}
}

func TestSaveToAndLoadFrom(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// Create config with values
	sopVersion := "0.7.0"
	goVersion := "1.23.0"
	config := Config{
		DefaultSop: (&sopVersion),
		DefaultGo: (&goVersion),
	}

	// Save
	_err0 := config.SaveTo(path)
	if _err0 != nil {
		err := _err0
		t.Fatalf("SaveTo failed: %v", err)
	}

	// Load back
	loaded, _err1 := LoadFrom(path)
	if _err1 != nil {
		err := _err1
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if loaded.DefaultSop == nil || (*loaded.DefaultSop) != sopVersion {
		t.Errorf("DefaultSop = %v, want %q", loaded.DefaultSop, sopVersion)
	}
	if loaded.DefaultGo == nil || (*loaded.DefaultGo) != goVersion {
		t.Errorf("DefaultGo = %v, want %q", loaded.DefaultGo, goVersion)
	}
}

func TestSaveToCreatesDirectory(t *testing.T) {
	dir := t.TempDir()

	// Nested path that doesn't exist
	path := filepath.Join(dir, "nested", "dir", "config.toml")

	config := Config{}
	_err0 := config.SaveTo(path)
	if _err0 != nil {
		err := _err0
		t.Fatalf("SaveTo failed: %v", err)
	}

	// Verify file exists
	_, _err1 := os.Stat(path)
	if _err1 != nil {
		err := _err1
		t.Errorf("config file not created: %v", err)
	}
}

func TestLoadProjectConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sop.mod")

	content := `sop = "0.5.0"
go = "1.22.0"
`
	_err0 := os.WriteFile(path, []byte(content), 0o644)
	if _err0 != nil {
		err := _err0
		t.Fatalf("failed to write sop.mod: %v", err)
	}

	config, _err1 := LoadProjectConfig(dir)
	if _err1 != nil {
		err := _err1
		t.Fatalf("LoadProjectConfig failed: %v", err)
	}

	if config.Sop == nil || (*config.Sop) != "0.5.0" {
		t.Errorf("Sop = %v, want 0.5.0", config.Sop)
	}
	if config.Go == nil || (*config.Go) != "1.22.0" {
		t.Errorf("Go = %v, want 1.22.0", config.Go)
	}
}

func TestLoadProjectConfigPartial(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sop.mod")

	// Only sop version specified
	content := `sop = "0.6.0"
`
	_err0 := os.WriteFile(path, []byte(content), 0o644)
	if _err0 != nil {
		err := _err0
		t.Fatalf("failed to write sop.mod: %v", err)
	}

	config, _err1 := LoadProjectConfig(dir)
	if _err1 != nil {
		err := _err1
		t.Fatalf("LoadProjectConfig failed: %v", err)
	}

	if config.Sop == nil || (*config.Sop) != "0.6.0" {
		t.Errorf("Sop = %v, want 0.6.0", config.Sop)
	}
	if config.Go != nil {
		t.Errorf("Go = %v, want nil", (*config.Go))
	}
}

func TestLoadProjectConfigNotFound(t *testing.T) {
	dir := t.TempDir()

	// Don't create sop.mod
	_, err := LoadProjectConfig(dir)
	if err == nil {
		t.Error("LoadProjectConfig should fail when sop.mod doesn't exist")
	}
}

