package main

import (
	"fmt"
	"net/http"
	"time"
	"log"

	"github.com/go-redis/redis/v8"

	"github.com/miniscruff/dashy/configs"
)

const timeFormat = time.ANSIC

func nowUTC() time.Time {
	return time.Now().UTC()
}

func (s *Server) CheckFeedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	feedName := r.URL.Path[15:]
	if feedName == "" {
		s.CheckAllFeeds()
		w.WriteHeader(http.StatusOK)
		return
	}

	feed := s.Config.FeedByName(feedName)
	if feed == nil {
		log.Printf("feed not found: '%v'\n", feedName)
		http.NotFound(w, r)
		return
	}

	err := s.CheckFeed(feed)
	if err != nil {
		http.Error(w, fmt.Errorf("unable to check feed: %w", err).Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) CheckAllFeeds() {
	log.Println("checking all feeds")
	for _, f := range s.Config.Feeds {
		s.CheckFeed(&f)
	}
}

func (s *Server) CheckFeed(feed *configs.FeedConfig) error {
	log.Printf("checking feed: %v\n", feed.Name)

	needsUpdate, err := s.feedOutOfDate(feed)
	if err != nil {
		return fmt.Errorf("unable to get feed time: %w", err)
	}

	if !needsUpdate {
		log.Printf("feed up to date: %v\n", feed.Name)
		return nil
	}

	s.UpdateFeed(feed)
	return nil
}

func (s *Server) feedOutOfDate(feed *configs.FeedConfig) (bool, error) {
	timeStr, err := s.RedisClient.Get(s.Ctx, configs.TimeKey(feed.Name)).Result()
	if err == redis.Nil || err != nil {
		return true, nil
	}

	lastRun, err := time.Parse(timeFormat, timeStr)
	if err != nil {
		return false, err
	}

	if feed.Schedule.Every != "" {
		return nowUTC().Sub(lastRun) > 0, nil
	}

	// on should also include a time zone
	// rename on to At actually...
	// if feed.Schedule.On != "" { }

	return false, nil
}
