package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/miniscruff/dashy/configs"
	"github.com/tidwall/gjson"
)

func (s *Server) UpdateFeedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	feedName := r.URL.Path[16:]

	feed := s.Config.FeedByName(feedName)
	if feed == nil {
		log.Printf("feed not found: '%v'\n", feedName)
		http.NotFound(w, r)
		return
	}

	err := s.UpdateFeed(feed)
	if err != nil {
		http.Error(w, fmt.Errorf("unable to update feed: %w", err).Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) UpdateFeed(feed *configs.FeedConfig) error {
	log.Printf("checking feed: %v\n", feed.Name)

	results, err := s.fetch(feed)
	if err != nil {
		return fmt.Errorf("unable to fetch data: %w", err)
	}

	err = s.storeResults(feed.Name, results)
	if err != nil {
		return fmt.Errorf("unable to store data: %w", err)
	}

	err = s.updateNextRun(feed)
	if err != nil {
		return fmt.Errorf("unable to update next run: %w", err)
	}

	log.Printf("feed updated: %v\n", feed.Name)
	return nil
}

func (s *Server) fetch(feed *configs.FeedConfig) (map[string]gjson.Result, error) {
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

func (s *Server) storeResults(name string, results map[string]gjson.Result) error {
	pipe := s.RedisClient.Pipeline()

	for k, result := range results {
		key := configs.ValueKey(name, k)
		if result.IsArray() {
			for _, v := range result.Array() {
				pipe.RPush(s.Ctx, key, v.Value())
			}
			start := -int64(len(result.Array()))
			pipe.LTrim(s.Ctx, key, start, -1)
		} else {
			pipe.Set(s.Ctx, key, result.Value(), 0)
		}
	}

	_, err := pipe.Exec(s.Ctx)
	return err
}

func (s *Server) updateNextRun(feed *configs.FeedConfig) error {
	dur, err := time.ParseDuration(feed.Schedule.Every)
	if err != nil {
		return err
	}

	nextTime := nowUTC().Add(dur).Format(timeFormat)
	_, err = s.RedisClient.Set(s.Ctx, configs.TimeKey(feed.Name), nextTime, 0).Result()
	return err
}
