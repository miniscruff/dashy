package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"

	"github.com/miniscruff/dashy/configs"
)

func main() {
	// ignore errors as .env may not exist
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
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

	tmpl := template.Must(template.New("view").Parse(config.View))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pipe := rdb.Pipeline()

		var keys []string
		for _, feed := range config.Feeds {
			for _, store := range feed.Store {
				key := configs.ValueKey(feed.Name, store.Name)
				if store.Count > 0 {
					pipe.LRange(ctx, key, 0, int64(store.Count))
				} else {
					pipe.Get(ctx, configs.ValueKey(feed.Name, store.Name))
				}
				keys = append(keys, feed.Name+strings.Title(store.Name))
			}
		}

		cmds, err := pipe.Exec(ctx)
		if err != nil {
			log.Println(fmt.Errorf("unable to get data: %w\n", err))
			http.Error(w, "unable to get data", 500)
			return
		}

		data := make(map[string]interface{}, 0)
		for i, cmd := range cmds {
			switch c := cmd.(type) {
			case *redis.StringCmd:
				data[keys[i]] = c.Val()
			case *redis.StringSliceCmd:
				data[keys[i]] = c.Val()
			default:
				fmt.Printf("unknown type: %v\n", c)
				data[keys[i]] = c.String()
			}
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			log.Println(fmt.Errorf("templating: %w", err))
		}
	})
	http.ListenAndServe(":"+port, nil)
}
