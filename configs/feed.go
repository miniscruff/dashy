package configs

import (
	"net/http"
	"net/url"
	"strings"
)

type FeedQuery struct {
	Headers map[string]string `yaml:"headers"`
	Params  map[string]string `yaml:"params"`
	Url     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Body    string            `yaml:"body"`
	Status  int               `yaml:"status"`
}

func (q FeedQuery) Request() (*http.Request, error) {
	bodyReader := strings.NewReader(q.Body)

	queryUrl := q.Url

	params := url.Values{}
	for k, v := range q.Params {
		params.Set(k, url.QueryEscape(stringOrEnvVar(v)))
	}

	if len(params) > 0 {
		queryUrl += "?" + params.Encode()
	}

	req, err := http.NewRequest(q.Method, queryUrl, bodyReader)
	if err != nil {
		return nil, err
	}

	for k, v := range q.Headers {
		req.Header.Add(k, stringOrEnvVar(v))
	}

	return req, err
}

type FeedSchedule struct {
	Every string `yaml:"every"`
	On    string `yaml:"on"`
}

type FeedStore struct {
	Name  string `yaml:"name"`
	Path  string `yaml:"path"`
	Count int    `yaml:"count"`
}

type FeedConfig struct {
	Name     string       `yaml:"name"`
	Query    FeedQuery    `yaml:"query"`
	Schedule FeedSchedule `yaml:"schedule"`
	Store    []FeedStore  `yaml:"store"`
}
