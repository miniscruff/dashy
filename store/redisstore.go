package store

import (
	"context"
	"errors"
	"log"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/tidwall/gjson"

	"github.com/miniscruff/dashy/configs"
)

func valueKey(name, value string) string {
	return fmt.Sprintf("value:%v:%v", name, value)
}

func timeKey(name string) string {
	return fmt.Sprintf("next-update:%v", name)
}

type RedisStore struct {
	config *configs.Config
	ctx    context.Context
	client *redis.Client
}

func NewRedisStore(config *configs.Config) (*RedisStore, error) {
	var (
		opts *redis.Options
		err  error
	)

	if config.Env.RedisUrl != "" {
		opts, err = redis.ParseURL(config.Env.RedisUrl)
		if err != nil {
			return nil, err
		}
	} else {
		if config.Env.RedisAddress == "" {
			return nil, errors.New("missing REDIS_ADDRESS env var")
		}

		opts = &redis.Options{
			Addr:     config.Env.RedisAddress,
			Username: config.Env.RedisUsername,
			Password: config.Env.RedisPassword,
			DB:       config.Env.RedisDatabase,
		}
	}

	client := redis.NewClient(opts)
	return &RedisStore{
		config: config,
		client: client,
		ctx:    context.Background(),
	}, nil
}

func (s *RedisStore) StringOrVar(value string) string {
	// add redis:KEY as well for things like refresh tokens
	if strings.HasPrefix(value, "env:") {
		return os.Getenv(value[4:])
	}
	return value
}

func (s *RedisStore) GetNextRun(feed *configs.FeedConfig) (time.Time, error) {
	timeStr, err := s.client.Get(s.ctx, timeKey(feed.Name)).Result()
	if err == redis.Nil || err != nil {
		return time.Time{}, nil
	}

	return time.Parse(timeFormat, timeStr)
}

func (s *RedisStore) SetNextRun(feed *configs.FeedConfig, nextRun time.Time) error {
	timeFormatted := nextRun.Format(timeFormat)
	_, err := s.client.Set(s.ctx, timeKey(feed.Name), timeFormatted, 0).Result()
	return err
}

func (s *RedisStore) GetValues() (map[string]map[string]interface{}, error) {
	pipe := s.client.Pipeline()

	var keys []string
	for _, feed := range s.config.Feeds {
		for _, store := range feed.Store {
			key := valueKey(feed.Name, store.Name)
			if store.Count > 0 {
				pipe.LRange(s.ctx, key, 0, int64(store.Count))
			} else {
				pipe.Get(s.ctx, valueKey(feed.Name, store.Name))
			}
			keys = append(keys, feed.Name+"|"+store.Name)
		}
	}

	cmds, err := pipe.Exec(s.ctx)
	if err != nil {
		log.Println(fmt.Errorf("unable to get data: %w\n", err))
		return nil, err
	}

	data := make(map[string]map[string]interface{}, 0)
	for i, cmd := range cmds {
		split := strings.Split(keys[i], "|")
		if _, exists := data[split[0]]; !exists {
			data[split[0]] = make(map[string]interface{}, 0)
		}

		switch c := cmd.(type) {
		case *redis.StringCmd:
			data[split[0]][split[1]] = c.Val()
		case *redis.StringSliceCmd:
			data[split[0]][split[1]] = c.Val()
		default:
			fmt.Printf("unknown type: %v\n", c)
			data[split[0]][split[1]] = c.String()
		}
	}

	return data, nil
}

func (s *RedisStore) SetValues(feed *configs.FeedConfig, values map[string]gjson.Result) error {
	pipe := s.client.Pipeline()

	for k, result := range values {
		key := valueKey(feed.Name, k)
		if result.IsArray() {
			for _, v := range result.Array() {
				pipe.RPush(s.ctx, key, v.Value())
			}
			start := -int64(len(result.Array()))
			pipe.LTrim(s.ctx, key, start, -1)
		} else {
			pipe.Set(s.ctx, key, result.Value(), 0)
		}
	}

	_, err := pipe.Exec(s.ctx)
	return err
}
