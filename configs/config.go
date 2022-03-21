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
	Feeds     []FeedConfig `yaml:"feeds"`
	Env       EnvConfig
	Dashboard Dashboard `yaml:"dashboard"`
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
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	IsArray bool   `yaml:"isArray,omitempty"`
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

/*
	Dashboard is the visual UI dashboard and its config
	It starts out at the base HTML level with meta and custom layouts / styles ( later )
	It also defines how many columns the grid should have

	Then you can define a number of layers:
		Each layer can have basic content or contents
*/

type Dashboard struct {
	Title        string            `yaml:"title"`
	Meta         map[string]string `yaml:"meta,omitempty"`
	CustomStyles map[string]string `yaml:"customStyles,omitempty"`
	Layers       []Layer           `yaml:"layers"`
}

type Layer struct {
	Name     string    `yaml:"name"`
	Title    string    `yaml:"title,omitempty"`
	X        int       `yaml:"x"`
	Y        int       `yaml:"y"`
	Width    int       `yaml:"width"`
	Height   int       `yaml:"height"`
	Layout   string    `yaml:"layout"`
	Contents []Content `yaml:"contents"`
}

type Content struct {
	Type   string   `yaml:"type"`
	Styles []string `yaml:"styles"`
	Text   string   `yaml:"text"`
}
