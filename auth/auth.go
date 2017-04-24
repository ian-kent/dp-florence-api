package auth

import (
	"context"
	"net/http"

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

			u, err := db.LoadUser(t)
			if err != nil {
				log.DebugR(req, "error authorising user", log.Data{"error": err})
				w.WriteHeader(401)
				return
			}

			log.DebugR(req, "user loaded", log.Data{"user": u})
			h.ServeHTTP(w, req.WithContext(withContext(req, t, &u)))
		})
	}
}

// // MustHaveRole ...
// func MustHaveRole(db *data.MongoDB, role string) func(h http.HandlerFunc) http.Handler {
// 	return func(h http.HandlerFunc) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
// 			u, ok := UserFromContext(req.Context())
// 			if !ok {
// 				w.WriteHeader(401)
// 				return
// 			}

// 			log.DebugR(req, "checking for role", log.Data{"role": role})

// 			var hasRole bool
// 			for _, r := range u.Roles {
// 				if r == role {
// 					hasRole = true
// 					break
// 				}
// 			}

// 			if !hasRole {
// 				log.DebugR(req, "role not found", log.Data{"role": role})
// 				w.WriteHeader(403)
// 				return
// 			}

// 			log.DebugR(req, "user has role", nil)
// 			h(w, req)
// 		})
// 	}
// }

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
