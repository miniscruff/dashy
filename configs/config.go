package configs

import (
	"gopkg.in/yaml.v2"
)

type Config struct {
	Timezone string       `yaml:"timezone"`
	Feeds    []FeedConfig `yaml:"feeds"`
	View     string       `yaml:"view"`
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
