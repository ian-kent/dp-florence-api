package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-florence-api/data"
	"github.com/ONSdigital/dp-florence-api/data/model"
	"github.com/ONSdigital/go-ns/log"
)

type ctxKey int

const (
	token ctxKey = iota
	user
)

// Middleware is the auth middleware
func Middleware(db *data.MongoDB) func(h http.HandlerFunc) http.Handler {
	return func(h http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			t := req.Header.Get("X-Florence-Token")
			log.DebugR(req, "auth", log.Data{"token": t})

			u, tok, err := db.LoadUserFromToken(t)
			if err != nil {
				log.DebugR(req, "error authorising user", log.Data{"error": err})
				w.WriteHeader(401)
				return
			}

			if tok.LastActive.Add(time.Minute * 60).Before(time.Now()) {
				log.DebugR(req, "token expired", nil)
				w.WriteHeader(401)
				return
			}

			err = db.UpdateTokenLastActive(t)
			if err != nil {
				log.ErrorR(req, err, nil)
				w.WriteHeader(500)
				return
			}

			log.DebugR(req, "user loaded", log.Data{"user": u})
			h.ServeHTTP(w, req.WithContext(withContext(req, t, &u)))
		})
	}
}

// WithPermission ...
func WithPermission(db *data.MongoDB, perm string) func(h http.HandlerFunc) http.Handler {
	return func(h http.HandlerFunc) http.Handler {
		return Middleware(db)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ok, err := HasPermission(req.Context(), db, perm)
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

			h.ServeHTTP(w, req)
		}))
	}
}

func withContext(req *http.Request, t string, u *model.User) context.Context {
	return context.WithValue(context.WithValue(req.Context(), token, t), user, u)
}

// UserFromContext ...
func UserFromContext(ctx context.Context) (u *model.User, ok bool) {
	u, ok = ctx.Value(user).(*model.User)
	return
}

// TokenFromContext ...
func TokenFromContext(ctx context.Context) (token string, ok bool) {
	u, ok := ctx.Value(token).(string)
	return u, ok
}

// HasPermission ...
func HasPermission(ctx context.Context, db *data.MongoDB, perm string) (ok bool, err error) {
	u, ok := UserFromContext(ctx)
	if !ok {
		return false, nil
	}

	for _, r := range u.Roles {
		role, err := db.GetRole(r)
		if err != nil {
			return false, err
		}

		if _, ok := role.Permissions[perm]; ok {
			return true, nil
		}
	}

	return false, nil
}
