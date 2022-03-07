package configs

import (
	_ "embed"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Feeds    []FeedConfig `yaml:"feeds"`
	View     string       `yaml:"view"`
}

func (c *Config) FeedByName(name string) *FeedConfig {
	for _, f := range c.Feeds {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

//go:embed config.yml
var configBytes []byte

func ReadConfig() (*Config, error) {
	var c   Config

	err := yaml.Unmarshal(configBytes, &c)
	if err != nil {
		return &c, err
	}

	return &c, nil
}
