package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/miniscruff/dashy/configs"
)

const timeFormat = time.ANSIC

func nowUTC() time.Time {
	return time.Now().UTC()
}

func redisTimeKey(feedName string) string {
	return fmt.Sprintf("next-update:%v", feedName)
}

func feedOutOfDate(ctx context.Context, feed *configs.FeedConfig, rdb *redis.Client) (bool, error) {
	timeStr, err := rdb.Get(ctx, redisTimeKey(feed.Name)).Result()
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
	_, err = rdb.Set(ctx, redisTimeKey(feed.Name), nextTime, 0).Result()
	return err
}
