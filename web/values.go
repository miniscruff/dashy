package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"

	"github.com/miniscruff/dashy/configs"
)

func (s *Server) ValuesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	pipe := s.RedisClient.Pipeline()

	var keys []string
	for _, feed := range s.Config.Feeds {
		for _, store := range feed.Store {
			key := configs.ValueKey(feed.Name, store.Name)
			if store.Count > 0 {
				pipe.LRange(s.Ctx, key, 0, int64(store.Count))
			} else {
				pipe.Get(s.Ctx, configs.ValueKey(feed.Name, store.Name))
			}
			keys = append(keys, feed.Name+"|"+store.Name)
		}
	}

	cmds, err := pipe.Exec(s.Ctx)
	if err != nil {
		log.Println(fmt.Errorf("unable to get data: %w\n", err))
		http.Error(w, "unable to get data", 500)
		return
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

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println(fmt.Errorf("unable to marshal data: %w\n", err))
		http.Error(w, "unable to marshal data", 500)
		return
	}

	w.Write(jsonData)
}
