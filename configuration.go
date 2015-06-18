package honeybadger

import (
	"log"
	"os"
	"time"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type Configuration struct {
	APIKey   string
	Root     string
	Env      string
	Hostname string
	Endpoint string
	Timeout  time.Duration
	Logger   Logger
	Backend  Backend
}

func (c1 Configuration) merge(c2 Configuration) Configuration {
	if c2.APIKey != "" {
		c1.APIKey = c2.APIKey
	}
	if c2.Root != "" {
		c1.Root = c2.Root
	}
	if c2.Env != "" {
		c1.Env = c2.Env
	}
	if c2.Hostname != "" {
		c1.Hostname = c2.Hostname
	}
	if c2.Endpoint != "" {
		c1.Endpoint = c2.Endpoint
	}
	if c2.Timeout > 0 {
		c1.Timeout = c2.Timeout
	}
	if c2.Logger != nil {
		c1.Logger = c2.Logger
	}
	if c2.Backend != nil {
		c1.Backend = c2.Backend
	}
	return c1
}

func newConfig(c Configuration) *Configuration {
	config := Configuration{
		APIKey:   getEnv("HONEYBADGER_API_KEY"),
		Root:     getPWD(),
		Env:      getEnv("HONEYBADGER_ENV"),
		Hostname: getHostname(),
		Endpoint: "https://api.honeybadger.io",
		Timeout:  3 * time.Second,
		Logger:   log.New(os.Stderr, "[honeybadger] ", log.Flags()),
	}.merge(c)

	if config.Backend == nil {
		config.Backend = newServerBackend(&config)
	}

	return &config
}

// Private helper methods
func getEnv(key string) string {
	return os.Getenv(key)
}

func getHostname() string {
	var hostname string
	hostname = getEnv("HONEYBADGER_HOSTNAME")
	if hostname == "" {
		if val, err := os.Hostname(); err == nil {
			hostname = val
		}
	}
	return hostname
}

func getPWD() string {
	var pwd string
	pwd = getEnv("HONEYBADGER_ROOT")
	if pwd == "" {
		if val, err := os.Getwd(); err == nil {
			pwd = val
		}
	}
	return pwd
}