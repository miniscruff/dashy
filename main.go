package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/miniscruff/dashy/configs"
	"github.com/miniscruff/dashy/server"
	"github.com/miniscruff/dashy/store"
)

var (
	//go:embed resources/static
	staticFS embed.FS

	//go:embed config.yml
	configBytes []byte
)

func main() {
	// ignore errors as .env may not exist
	_ = godotenv.Load()

	cfg, err := configs.NewConfig(configBytes)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to load config: %w", err))
	}

	redisStore, err := store.NewRedisStore(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to load redis: %w", err))
	}

	server := server.Server{
		StaticFS:  staticFS,
		Config:    cfg,
		Store:     redisStore,
	}
	if err := server.Serve(); err != nil {
		log.Fatal(fmt.Errorf("unable to serve: %w", err))
	}
}
