package server

import (
	"embed"
	"fmt"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/miniscruff/dashy/configs"
	"github.com/miniscruff/dashy/store"
)

type Server struct {
	StaticFS  embed.FS
	IndexFile []byte
	Config    *configs.Config
	Store     store.Store
}

func (s *Server) StaticFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	name := "resources/static/" + r.URL.Path[8:]
	fBytes, err := s.StaticFS.ReadFile(name)
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

	w.Write(s.IndexFile)
}

func (s *Server) Serve() error {
	// hook up handlers
	http.HandleFunc("/api/checkFeeds", s.CheckFeedHandler)
	http.HandleFunc("/api/checkFeed/", s.CheckFeedHandler)
	http.HandleFunc("/api/updateFeed/", s.UpdateFeedHandler)
	http.HandleFunc("/api/values", s.ValuesHandler)
	http.HandleFunc("/api/events", s.EventsHandler)
	http.HandleFunc("/static/", s.StaticFileHandler)
	http.HandleFunc("/", s.IndexHandler)

	// auto tick every 5 minutes
	ticker := time.NewTicker(s.Config.Env.TickDuration)
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

	host := fmt.Sprintf("%v:%v", s.Config.Env.Address, s.Config.Env.Port)
	log.Println("listening on", host)
	http.ListenAndServe(host, nil)
	return nil
}
