package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/tidwall/gjson"

	"github.com/miniscruff/dashy/configs"
)

func fetch(feed *configs.FeedConfig) (map[string]gjson.Result, error) {
	// might need configs for clients later...
	client := &http.Client{}

	req, err := feed.Query.Request()
	if err != nil {
		return nil, fmt.Errorf("unable to create request from query: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to get response: %w", err)
	}

	if res.StatusCode != feed.Query.Status {
		return nil, fmt.Errorf("status code '%v' does not match expected '%v'", res.StatusCode, feed.Query.Status)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response bytes: %w", err)
	}
	defer res.Body.Close()

	if !gjson.ValidBytes(bodyBytes) {
		return nil, errors.New("body is not a valid JSON")
	}

	var paths []string
	for _, s := range feed.Store {
		paths = append(paths, s.Path)
	}

	jsonResults := gjson.GetManyBytes(bodyBytes, paths...)

	results := make(map[string]gjson.Result, 0)
	for i, s := range feed.Store {
		results[s.Name] = jsonResults[i]
	}

	return results, nil
}
