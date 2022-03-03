package main

import (
	"context"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/go-redis/redis/v8"
)

func storeResults(ctx context.Context, name string, results map[string]gjson.Result, rdb *redis.Client) error {
	pipe := rdb.Pipeline()

	for k, result := range results {
		key := fmt.Sprintf("value:%v:%v", name, k)
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
