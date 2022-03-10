package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (s *Server) ValuesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	data, err := s.Store.GetValues()
	if err != nil {
		log.Println(fmt.Errorf("unable to get data: %w\n", err))
		http.Error(w, "unable to get data", 500)
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println(fmt.Errorf("unable to marshal data: %w\n", err))
		http.Error(w, "unable to marshal data", 500)
		return
	}

	w.Write(jsonData)
}
