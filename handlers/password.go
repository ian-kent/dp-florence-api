package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-florence-api/data"
	"github.com/ONSdigital/go-ns/log"
)

type passwordInput struct {
	Email       string `json:"email"`
	OldPassword string `json:"oldPassword"`
	Password    string `json:"password"`
}

// ChangePassword ...
func (s *FloServer) ChangePassword(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.DebugR(req, "error reading body", log.Data{"error": err})
		w.WriteHeader(400)
		return
	}

	var input passwordInput
	err = json.Unmarshal(b, &input)
	if err != nil {
		log.DebugR(req, "error unmarshaling data", log.Data{"error": err})
		w.WriteHeader(400)
		return
	}

	err = s.DB.ChangePassword(input.Email, input.OldPassword, input.Password)
	if err != nil {
		log.DebugR(req, "error changing password", log.Data{"error": err})

		if err == data.ErrUserNotFound || err == data.ErrUserInactive {
			w.WriteHeader(400)
			return
		} else if err == data.ErrInvalidPassword {
			w.WriteHeader(401)
			return
		}

		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(`"Password updated for ` + input.Email + `"`))
}
