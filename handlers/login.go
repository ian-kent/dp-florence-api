package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-florence-api/data"
	"github.com/ONSdigital/go-ns/log"
)

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login ...
func (s *FloServer) Login(w http.ResponseWriter, req *http.Request) {
	var input loginInput

	if err := unmarshal(req, &input); err != nil {
		log.DebugR(req, "error reading body", log.Data{"error": err})
		w.WriteHeader(400)
		return
	}

	token, err := s.DB.ValidateLogin(input.Email, input.Password)
	if err != nil {
		log.DebugR(req, "invalid username or password", log.Data{"error": err})

		if err == data.ErrForcePasswordChange {
			w.WriteHeader(417)
			return
		}

		w.WriteHeader(401)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write([]byte(`"` + token + `"`))
	if err != nil {
		log.DebugR(req, "error writing response", log.Data{"error": err})
	}
}
