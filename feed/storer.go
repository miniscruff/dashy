package main

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/miniscruff/dashy/configs"
	"github.com/tidwall/gjson"
)

func storeResults(ctx context.Context, name string, results map[string]gjson.Result, rdb *redis.Client) error {
	pipe := rdb.Pipeline()

	for k, result := range results {
		key := configs.ValueKey(name, k)
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

	_, err := pipe.Exec(ctx)
	return err
}