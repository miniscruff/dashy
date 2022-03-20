package configs

import (
	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v2"
	"time"
)

func NewConfig(configYml []byte) (*Config, error) {
	var c Config

	err := yaml.Unmarshal(configYml, &c)
	if err != nil {
		return &c, err
	}

	if err := env.Parse(&c.Env); err != nil {
		return &c, err
	}

	return &c, nil
}

type Config struct {
	Feeds []FeedConfig `yaml:"feeds"`
	Env   EnvConfig
}

type FeedConfig struct {
	Name     string       `yaml:"name"`
	Query    FeedQuery    `yaml:"query"`
	Schedule FeedSchedule `yaml:"schedule"`
	Store    []FeedStore  `yaml:"store"`
}

type FeedQuery struct {
	Headers map[string]string `yaml:"headers"`
	Params  map[string]string `yaml:"params"`
	Url     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Body    string            `yaml:"body"`
	Status  int               `yaml:"status"`
}

type FeedSchedule struct {
	Every string `yaml:"every"`
}

type FeedStore struct {
	Name  string `yaml:"name"`
	Path  string `yaml:"path"`
	Count int    `yaml:"count,omitempty"`
}

func (c *Config) FeedByName(name string) *FeedConfig {
	for _, f := range c.Feeds {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

// EnvConfig contains values expected from the environment
type EnvConfig struct {
	RedisUrl      string `env:"REDIS_URL"`
	RedisAddress  string `env:"REDIS_ADDRESS"`
	RedisUsername string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisDatabase int    `env:"REDIS_DATABASE"`

	Port         string        `env:"PORT" envDefault:"8080"`
	Address      string        `env:"ADDRESS" envDefault:"0.0.0.0"`
	TickDuration time.Duration `env:"TICK_DURATION" envDefault:"15m"`
}
