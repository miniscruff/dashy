package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/joho/godotenv"
	"github.com/go-redis/redis/v8"

	"github.com/miniscruff/dashy/configs"
)

func checkFeed(ctx context.Context, name string, config *configs.Config, rdb *redis.Client) error {
	log.Printf("checking feed: %v\n", name)

	feed := config.FeedByName(name)
	if feed == nil {
		return fmt.Errorf("feed not found: '%v'", name)
	}

	needsUpdate, err := feedOutOfDate(ctx, feed, rdb)
	if err != nil {
		return fmt.Errorf("unable to get feed time: %w", err)
	}

	if !needsUpdate {
		log.Printf("feed up to date: %v\n", name)
		return nil
	}

	log.Printf("feed scheduled for update: %v\n", name)
	rdb.Publish(ctx, "updater", name)
	return nil
}

func updateFeed(ctx context.Context, name string, config *configs.Config, rdb *redis.Client) error {
	log.Printf("updating feed: %v\n", name)

	feed := config.FeedByName(name)
	if feed == nil {
		return fmt.Errorf("feed not found: '%v'", name)
	}

	results, err := fetch(feed)
	if err != nil {
		return fmt.Errorf("unable to fetch data: %w", err)
	}

	err = storeResults(ctx, feed.Name, results, rdb)
	if err != nil {
		return fmt.Errorf("unable to store data: %w", err)
	}

	err = updateNextRun(ctx, feed, rdb)
	if err != nil {
		return fmt.Errorf("unable to update next run: %w", err)
	}

	log.Printf("feed updated: %v\n", name)
	return nil
}

func main() {
	// ignore errors as .env may not exist
	_ = godotenv.Load()

	config, err := configs.ReadConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("reading config: %w", err))
	}

	rdb, err := configs.RedisClient()
	if err != nil {
		log.Fatal(fmt.Errorf("connecting to redis: %w", err))
	}

	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Add(1)

	go func() {
		scheduleSub := rdb.Subscribe(ctx, "scheduler")

		for {
			msg, err := scheduleSub.ReceiveMessage(ctx)
			if err != nil {
				log.Println(fmt.Errorf("receiving message: %w", err))
				break
			}

			name := msg.Payload
			err = checkFeed(ctx, name, config, rdb)
			if err != nil {
				log.Println(fmt.Errorf("unable to check feed '%v': %w", name, err))
			}
		}
		wg.Done()
	}()

	go func() {
		scheduleSub := rdb.Subscribe(ctx, "updater")

		for {
			msg, err := scheduleSub.ReceiveMessage(ctx)
			if err != nil {
				log.Println(fmt.Errorf("error: %w", err))
				break
			}

			name := msg.Payload
			err = updateFeed(ctx, name, config, rdb)
			if err != nil {
				log.Println(fmt.Errorf("unable to update feed '%v': %w", name, err))
			}
		}
		wg.Done()
	}()

	wg.Wait()
}
