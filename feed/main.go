package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/tidwall/gjson"

	"github.com/miniscruff/dashy/configs"
)

// value : <feed name> : <store name>
const saveFormat = "value:%v:%v"

func redisOptions() (*redis.Options, error) {
	if url, ok := os.LookupEnv("REDIS_URL"); ok {
		return redis.ParseURL(url)
	}

	addr, ok := os.LookupEnv("REDIS_ADDRESS")
	if !ok {
		return nil, errors.New("missing REDIS_ADDRESS env var")
	}

	user := os.Getenv("REDIS_USERNAME")
	pass := os.Getenv("REDIS_PASSWORD")

	db := 0

	if dbVar, ok := os.LookupEnv("REDIS_DB"); ok {
		var err error
		db, err = strconv.Atoi(dbVar)
		if err != nil {
			return nil, err
		}
	}

	return &redis.Options{
		Addr: addr,
		Username: user,
		Password: pass,
		DB: db,
	}, nil
}

func main() {
	feedName := os.Args[1]

	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Errorf("Error loading .env file: %w", err))
	}

	ctx := context.Background()
	opts, err := redisOptions()
	if err != nil {
		log.Fatal(fmt.Errorf("loading redis options: %w", err))
	}

	rdb := redis.NewClient(opts)

	config, err := configs.ReadConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("reading config: %w", err))
	}

	// might need configs for clients later...
	client := &http.Client{}

	feed := config.FeedByName(feedName)
	if feed == nil {
		log.Fatalf("feed by name '%v' not found in config", feedName)
	}

	req, err := feed.Query.Request()
	if err != nil {
		log.Fatal(fmt.Errorf("unable to create request from query: %w", err))
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to get response: %w", err))
	}

	if res.StatusCode != feed.Query.Status {
		log.Fatalf("status code '%v' does not match expected '%v'", res.StatusCode, feed.Query.Status)
	}

	// get results from body
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to read response bytes: %w", err))
	}
	defer res.Body.Close()

	results, err := getResults(bodyBytes, feed.Store, config)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to get values from body: %w", err))
	}

	// save results...
	pipe := rdb.Pipeline()

	for k, result := range results {
		key := fmt.Sprintf(saveFormat, feedName, k)
		if result.IsArray() {
			for _, v := range result.Array() {
				pipe.RPush(ctx, key, v.Value())
			}
			start := -int64(len(result.Array()))
			pipe.LTrim(ctx, key, start, -1)
		} else {
			pipe.Set(ctx, key, result.Value(), 0)
		}
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to save results: %w", err))
	}
}

func getResults(body []byte, store []configs.FeedStore, config *configs.Config) (map[string]gjson.Result, error) {
	results := make(map[string]gjson.Result, 0)

	if !gjson.ValidBytes(body) {
		return nil, errors.New("body is not a valid JSON")
	}

	var paths []string
	for _, s := range store {
		paths = append(paths, s.Path)
	}

	res := gjson.GetManyBytes(body, paths...)
	for i, s := range store {
		results[s.Name] = res[i]
	}

	return results, nil
}

func writeResults() error {
	return nil
}
