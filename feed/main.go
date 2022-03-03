package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/miniscruff/dashy/configs"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Errorf("Error loading .env file: %w", err))
	}

	config, err := configs.ReadConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("reading config: %w", err))
	}

	rdb, err := configs.RedisClient()
	if err != nil {
		log.Fatal(fmt.Errorf("connecting to redis: %w", err))
	}

	ctx := context.Background()

	for _, feed := range config.Feeds {
		fmt.Printf("checking feed: %v\n", feed.Name)

		needsUpdate, err := feedOutOfDate(ctx, &feed, rdb)
		if err != nil {
			log.Fatal(fmt.Errorf("unable to get feed time: %w", err))
		}

		if needsUpdate {
			fmt.Println("out of date")

			results, err := fetch(&feed, config)
			if err != nil {
				log.Fatal(fmt.Errorf("unable to fetch data: %w", err))
			}

			err = storeResults(ctx, feed.Name, results, rdb)
			if err != nil {
				log.Fatal(fmt.Errorf("unable to store data: %w", err))
			}

			err = updateNextRun(ctx, &feed, rdb)
			if err != nil {
				log.Fatal(fmt.Errorf("unable to update next run: %w", err))
			}

			fmt.Println("feed updated")
		}
	}
}
