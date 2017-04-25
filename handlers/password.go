package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-florence-api/auth"
	"github.com/ONSdigital/dp-florence-api/data"
	"github.com/ONSdigital/dp-florence-api/data/model"
	"github.com/ONSdigital/go-ns/log"
)

type passwordInput struct {
	Email       string `json:"email"`
	OldPassword string `json:"oldPassword"`
	Password    string `json:"password"`
}

// ChangePassword ...
func (s *FloServer) ChangePassword(w http.ResponseWriter, req *http.Request) {
	var input passwordInput
	if err := unmarshal(req, &input); err != nil {
		log.DebugR(req, "error reading body", log.Data{"error": err})
		w.WriteHeader(400)
		return
	}

	u, ok := auth.UserFromContext(req.Context())
	adminOK, err := auth.HasPermission(req.Context(), s.DB, model.PermAdministrator)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}
	log.DebugR(req, "change password", log.Data{
		"ok":         ok,
		"adminOK":    ok,
		"inputEmail": input.Email,
		"uEmail":     u.Email,
	})
	if ok && adminOK && input.Email != u.Email {
		// In old-Zebedee, this would update a users password
		// Now it silently returns an OK response to allow create user to complete
		// But really it does nothing, since we now send a verification email instead
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
		return
	}

	if err := s.DB.ChangePassword(input.Email, input.OldPassword, input.Password); err != nil {
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
