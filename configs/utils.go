package configs

import (
	"errors"
	"strings"
	"strconv"
	"os"

	"github.com/go-redis/redis/v8"
)

// add redis:KEY as well for things like refresh tokens
func stringOrEnvVar(value string) string {
	if strings.HasPrefix(value, "env:") {
		return os.Getenv(value[4:])
	}
	return value
}

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

func RedisClient() (*redis.Client, error) {
	opts, err := redisOptions()
	if err != nil {
		return nil, err
	}

	return redis.NewClient(opts), nil
}
