package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-florence-api/auth"
	"github.com/ONSdigital/dp-florence-api/data"
	"github.com/ONSdigital/dp-florence-api/data/model"
	"github.com/ONSdigital/go-ns/log"
)

type permissionsOutput struct {
	Email            string `json:"email"`
	Admin            bool   `json:"admin"`
	DataVisPublisher bool   `json:"dataVisPublisher"`
	Editor           bool   `json:"editor"`

	Roles []roleOutput `json:"roles"`
}

type roleOutput struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	Permissions map[string]permissionOutput `json:"permissions"`
}

type permissionOutput map[string]interface{}

type permissionsInput struct {
	Email            string `json:"email"`
	Admin            bool   `json:"admin"`
	Editor           bool   `json:"editor"`
	DataVisPublisher bool   `json:"dataVisPublisher"`
}

// GetPermissions ...
func (s *FloServer) GetPermissions(w http.ResponseWriter, req *http.Request) {
	email := req.URL.Query().Get("email")
	if len(email) == 0 {
		w.WriteHeader(400)
		return
	}

	u, err := s.DB.GetUser(email)
	if err != nil {
		if err == data.ErrUserNotFound {
			w.WriteHeader(404)
			return
		}
		log.ErrorR(req, err, log.Data{})
		w.WriteHeader(500)
		return
	}

	var p permissionsOutput
	p.Email = email

	for _, r := range u.Roles {
		role, err := s.DB.GetRole(r)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(500)
			return
		}

		rO := roleOutput{
			ID:          role.ID,
			Name:        role.Name,
			Permissions: make(map[string]permissionOutput),
		}
		for k := range role.Permissions {
			if k == model.PermAdministrator {
				p.Admin = true
			} else if k == model.PermEditor {
				// FIXME handle data vis permission
				p.Editor = true
			}
			rO.Permissions[k] = permissionOutput{}
		}
		p.Roles = append(p.Roles, rO)
	}

	b, err := json.Marshal(&p)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(b)
	if err != nil {
		log.DebugR(req, "error writing response", log.Data{"error": err})
		return
	}
}

// UpdatePermissions ...
func (s *FloServer) UpdatePermissions(w http.ResponseWriter, req *http.Request) {
	creator, ok := auth.UserFromContext(req.Context())
	if !ok {
		log.DebugR(req, "user not logged in", nil)
		w.WriteHeader(401)
		return
	}

	var input permissionsInput
	if err := unmarshal(req, &input); err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(400)
		return
	}

	var roles []string

	if input.Admin {
		roles = append(roles, "administrator")
	}
	if input.Editor {
		roles = append(roles, "editor")
	}
	// FIXME handle data vis users

	err := s.DB.SetUserRoles(creator.ID.Hex(), input.Email, roles...)
	if err != nil {
		log.ErrorR(req, err, nil)
		if err == data.ErrUserExists {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(`{}`))
}
