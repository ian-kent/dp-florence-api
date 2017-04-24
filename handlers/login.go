package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-florence-api/data"
	"github.com/ONSdigital/go-ns/log"
)

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginOutput struct {
	Token string `json:"token"`
}

// Login ...
func (s *FloServer) Login(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.DebugR(req, "error reading body", log.Data{"error": err})
		w.WriteHeader(400)
		return
	}

	var input loginInput
	err = json.Unmarshal(b, &input)
	if err != nil {
		log.DebugR(req, "error unmarshaling data", log.Data{"error": err})
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

	// var output loginOutput
	// output.Token = token

	// b, err = json.Marshal(&output)
	// if err != nil {
	// 	log.DebugR(req, "error marshaling data", log.Data{"error": err})
	// 	w.WriteHeader(500)
	// 	return
	// }

	_, err = w.Write([]byte(`"` + token + `"`))
	if err != nil {
		log.DebugR(req, "error writing response", log.Data{"error": err})
	}
}
