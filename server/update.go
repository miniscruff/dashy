package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
	"strings"

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
	log.Printf("updating feed: %v\n", feed.Name)

	results, err := s.fetch(feed)
	if err != nil {
		return fmt.Errorf("unable to fetch data: %w", err)
	}

	err = s.Store.SetValues(feed.Name, results)
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

func (s *Server) request(query *configs.FeedQuery) (*http.Request, error) {
	bodyReader := strings.NewReader(query.Body)

	queryUrl := query.Url

	params := url.Values{}
	for k, v := range query.Params {
		params.Set(k, url.QueryEscape(s.Store.StringOrVar(v)))
	}

	if len(params) > 0 {
		queryUrl += "?" + params.Encode()
	}

	req, err := http.NewRequest(query.Method, queryUrl, bodyReader)
	if err != nil {
		return nil, err
	}

	for k, v := range query.Headers {
		req.Header.Add(k, s.Store.StringOrVar(v))
	}

	return req, err
}

func (s *Server) fetch(feed *configs.FeedConfig) (map[string]gjson.Result, error) {
	// might need configs for clients later...
	client := &http.Client{}

	req, err := s.request(&feed.Query)
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

func (s *Server) updateNextRun(feed *configs.FeedConfig) error {
	dur, err := time.ParseDuration(feed.Schedule.Every)
	if err != nil {
		return err
	}

	nextTime := time.Now().UTC().Add(dur)
	return s.Store.SetNextRun(feed.Name, nextTime)
}
