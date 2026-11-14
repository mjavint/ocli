package config

import "sync"

// Loader loads the ocli configuration lazily and caches the result.
type Loader struct {
	configFile string
	once       sync.Once
	config     *Config
}

// NewLoader creates a lazy configuration loader.
func NewLoader(configFile string) *Loader {
	return &Loader{configFile: configFile}
}

// Load returns the configuration, falling back to defaults on error.
func (l *Loader) Load() *Config {
	l.once.Do(func() {
		manager := NewConfigManager(l.configFile)
		if err := manager.Load(); err != nil {
			l.config = manager.getDefaultConfig()
			return
		}
		if manager.config != nil {
			l.config = manager.config
			return
		}
		l.config = manager.getDefaultConfig()
	})
	return l.config
}
