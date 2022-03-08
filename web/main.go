package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"

	"github.com/miniscruff/dashy/configs"
)

//go:embed resources/static
var staticFS embed.FS

//go:embed resources/index.html
var indexFile []byte

type Server struct {
	Ctx         context.Context
	Config      *configs.Config
	RedisClient *redis.Client
	// Logger later...
	// custom logger that writes "errors" to redis as a list
	// then we can show errors as part of a UI panel

	address      string
	port         string
	tickDuration time.Duration
}

func envOrDefault(key, def string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}

	return def
}

func NewServer() (*Server, error) {
	// ignore errors as .env may not exist
	_ = godotenv.Load()

	port := envOrDefault("PORT", "5000")
	address := envOrDefault("ADDRESS", "localhost")
	tickDuration := envOrDefault("TICK_DURATION", "5m")
	duration, err := time.ParseDuration(tickDuration)
	if err != nil {
		return nil, fmt.Errorf("invalid tick duration '%v': %w", tickDuration, err)
	}

	config, err := configs.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	rdb, err := configs.RedisClient()
	if err != nil {
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}

	ctx := context.Background()

	return &Server{
		Ctx:          ctx,
		Config:       config,
		RedisClient:  rdb,
		address:      address,
		port:         port,
		tickDuration: duration,
	}, nil
}

func (s *Server) StaticFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	name := "resources/static/" + r.URL.Path[8:]
	fBytes, err := staticFS.ReadFile(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ctype := mime.TypeByExtension(filepath.Ext(name))
	w.Header().Add("Content-Type", ctype)
	w.Write(fBytes)
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" || r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	w.Write(indexFile)
}

func (s *Server) Serve() {
	// hook up handlers
	http.HandleFunc("/api/checkFeeds", s.CheckFeedHandler)
	http.HandleFunc("/api/checkFeed/", s.CheckFeedHandler)
	http.HandleFunc("/api/updateFeed/", s.UpdateFeedHandler)
	http.HandleFunc("/api/values", s.ValuesHandler)
	http.HandleFunc("/api/events", s.EventsHandler)
	http.HandleFunc("/static/", s.StaticFileHandler)
	http.HandleFunc("/", s.IndexHandler)

	// auto tick every 5 minutes
	ticker := time.NewTicker(s.tickDuration)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.CheckAllFeeds()
			}
		}
	}()
	// run at startup as well
	go s.CheckAllFeeds()

	log.Printf("listening on %v:%v\n", s.address, s.port)
	http.ListenAndServe(fmt.Sprintf("%v:%v", s.address, s.port), nil)
}

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal(fmt.Errorf("unable to initialize server: %w", err))
	}

	server.Serve()
}
