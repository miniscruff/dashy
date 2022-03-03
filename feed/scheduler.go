package main

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/miniscruff/dashy/configs"
)

const timeFormat = time.ANSIC

func nowUTC() time.Time {
	return time.Now().UTC()
}

func feedOutOfDate(ctx context.Context, feed *configs.FeedConfig, rdb *redis.Client) (bool, error) {
	timeStr, err := rdb.Get(ctx, configs.TimeKey(feed.Name)).Result()
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

func updateNextRun(ctx context.Context, feed *configs.FeedConfig, rdb *redis.Client) error {
	dur, err := time.ParseDuration(feed.Schedule.Every)
	if err != nil {
		return err
	}

	nextTime := nowUTC().Add(dur).Format(timeFormat)
	_, err = rdb.Set(ctx, configs.TimeKey(feed.Name), nextTime, 0).Result()
	return err
}
