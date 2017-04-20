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
			t := req.Header.Get("Authorization")
			log.DebugR(req, "auth", log.Data{"token": t})

			u, err := db.LoadUser(t)
			if err != nil {
				log.DebugR(req, "error authorising user", log.Data{"error": err})
				w.WriteHeader(401)
				return
			}

			log.DebugR(req, "user loaded", log.Data{"user": u})
			h.ServeHTTP(w, req.WithContext(withContext(t, u)))
		})
	}
}

func withContext(t string, u *model.User) context.Context {
	return context.WithValue(context.WithValue(context.Background(), token, t), user, u)
}

// UserFromContext ...
func UserFromContext(ctx context.Context) (user *model.User, ok bool) {
	u, ok := ctx.Value(user).(*model.User)
	return u, ok
}

// TokenFromContext ...
func TokenFromContext(ctx context.Context) (token string, ok bool) {
	u, ok := ctx.Value(token).(string)
	return u, ok
}
