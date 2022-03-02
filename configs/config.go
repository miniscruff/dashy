package configs

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Timezone string       `yaml:"timezone"`
	Feeds    []FeedConfig `yaml:"feeds"`
}

func (c *Config) FeedByName(name string) *FeedConfig {
	for _, f := range c.Feeds {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

const configEnvVar = "DASHY_CONFIG_PATH"
const configPath = "config.yml"

func ReadConfig() (*Config, error) {
	var (
		c   Config
		bs  []byte
		err error
	)

	var path string
	if os.Getenv(configEnvVar) != "" {
		path = os.Getenv(configEnvVar)
	} else {
		path = configPath
	}

	bs, err = os.ReadFile(path)
	if err != nil {
		return &c, err
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return &c, err
	}

	return &c, nil
}
