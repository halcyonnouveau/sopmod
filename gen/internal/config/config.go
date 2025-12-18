//soppo:generated
package config

import "os"
import "path/filepath"
import "github.com/BurntSushi/toml"
import "github.com/halcyonnouveau/sopmod/gen/internal/paths"

// Config is the global sopmod configuration stored in ~/.sopmod/config.toml
type Config struct {
    DefaultSop *string `toml:"default_sop,omitempty"`
    DefaultGo *string `toml:"default_go,omitempty"`
}

// Load loads config from ~/.sopmod/config.toml, or returns default if not found
func Load() Config {
    path := paths.ConfigPath()
    config, _err0 := LoadFrom(path)
    if _err0 != nil {
        return Config{}
    }
    return config
}

// LoadFrom loads config from a specific path
func LoadFrom(path string) (Config, error) {
    var config Config
    _, _err0 := toml.DecodeFile(path, (&config))
    if _err0 != nil {
        return Config{}, _err0
    }
    return config, nil
}

// Save saves config to ~/.sopmod/config.toml
func (c *Config) Save() error {
    path := paths.ConfigPath()
    return c.SaveTo(path)
}

// SaveTo saves config to a specific path
func (c *Config) SaveTo(path string) error {
    dir := filepath.Dir(path)
    _err0 := os.MkdirAll(dir, 493)
    if _err0 != nil {
        return _err0
    }
    f, _err1 := os.Create(path)
    if _err1 != nil {
        return _err1
    }
    defer f.Close()

    encoder := toml.NewEncoder(f)
    return encoder.Encode(c)
}

// ProjectConfig holds project-specific version requirements from sop.mod
type ProjectConfig struct {
    Go *string `toml:"go,omitempty"`
    Sop *string `toml:"sop,omitempty"`
}

// LoadProjectConfig loads project config from sop.mod in the given directory
func LoadProjectConfig(dir string) (*ProjectConfig, error) {
    path := filepath.Join(dir, "sop.mod")
    var config ProjectConfig
    _, _err0 := toml.DecodeFile(path, (&config))
    if _err0 != nil {
        return nil, _err0
    }
    return (&config), nil
}

