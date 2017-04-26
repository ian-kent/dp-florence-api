package handlers

import (
	"net/http"
	"strings"

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

	if strings.HasPrefix(input.Email, "<verify>:") {
		// TODO this is a nasty hack, Florence could be refactored properly
		// to handle user verification in a nicer way!
		log.DebugR(req, "user verification", log.Data{"token": input.Password})
		ok, err := s.DB.ValidateUserVerificationCode(input.Password)
		if err != nil || ok != true {
			log.DebugR(req, "error validating code", log.Data{"error": err})
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(417)
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
