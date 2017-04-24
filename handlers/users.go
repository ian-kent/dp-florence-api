package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-florence-api/data"
	"github.com/ONSdigital/go-ns/log"
)

type userOutput struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	Inactive          bool   `json:"inactive"`
	TemporaryPassword bool   `json:"temporaryPassword"`
	LastAdmin         string `json:"lastAdmin"`
}

// ListUsers ...
func (s *FloServer) ListUsers(w http.ResponseWriter, req *http.Request) {
	// FIXME this should be a different URL!
	if len(req.URL.Query().Get("email")) > 0 {
		s.GetUser(w, req)
		return
	}

	u := []userOutput{}

	users, err := s.DB.GetUsers()
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	for _, user := range users {
		u = append(u, userOutput{
			Name:              user.Name,
			Email:             user.Email,
			Inactive:          !user.Active,
			TemporaryPassword: user.ForcePasswordChange,
			LastAdmin:         "", // FIXME this seems useless?
		})
	}

	b, err := json.Marshal(&u)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// GetUser ...
func (s *FloServer) GetUser(w http.ResponseWriter, req *http.Request) {
	email := req.URL.Query().Get("email")

	user, err := s.DB.GetUser(email)
	if err != nil {
		log.ErrorR(req, err, nil)
		if err == data.ErrUserNotFound {
			w.WriteHeader(404)
			return
		}

		w.WriteHeader(500)
		return
	}

	u := userOutput{
		Name:              user.Name,
		Email:             user.Email,
		Inactive:          !user.Active,
		TemporaryPassword: user.ForcePasswordChange,
		LastAdmin:         "", // FIXME this seems useless?
	}

	b, err := json.Marshal(&u)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
