package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-florence-api/auth"
	"github.com/ONSdigital/dp-florence-api/data"
	"github.com/ONSdigital/dp-florence-api/data/model"
	"github.com/ONSdigital/go-ns/log"
)

type userOutput struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	Inactive          bool   `json:"inactive"`
	TemporaryPassword bool   `json:"temporaryPassword"`
	LastAdmin         string `json:"lastAdmin"`
}

type createUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
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

// CreateUser ...
func (s *FloServer) CreateUser(w http.ResponseWriter, req *http.Request) {
	creator, ok := auth.UserFromContext(req.Context())
	if !ok {
		log.DebugR(req, "user not logged in", nil)
		w.WriteHeader(401)
		return
	}

	ok, err := auth.HasPermission(req.Context(), s.DB, model.PermAdministrator)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	if !ok {
		log.DebugR(req, "user needs administrator permission", nil)
		w.WriteHeader(403)
		return
	}

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(400)
		return
	}

	var input createUserInput
	err = json.Unmarshal(b, &input)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(400)
		return
	}

	err = s.DB.CreateUser(creator.ID.Hex(), input.Email, input.Name)
	if err != nil {
		log.ErrorR(req, err, nil)
		if err == data.ErrUserExists {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(201)
}
