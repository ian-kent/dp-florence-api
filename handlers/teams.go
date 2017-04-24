package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
)

type teamsOutput struct {
	Teams []teamOutput `json:"teams"`
}

type teamOutput struct {
}

// ListTeams ...
func (s *FloServer) ListTeams(w http.ResponseWriter, req *http.Request) {
	t := teamsOutput{
		Teams: make([]teamOutput, 0),
	}

	b, err := json.Marshal(&t)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
