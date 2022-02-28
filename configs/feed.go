package configs

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
	Count int    `yaml:"count"`
	Type  string `yaml:"type"`
}

type FeedConfig struct {
	Name     string       `yaml:"name"`
	Query    FeedQuery    `yaml:"query"`
	Schedule FeedSchedule `yaml:"schedule"`
	Store    []FeedStore  `yaml:"store"`
}
