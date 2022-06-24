package middleware

import (
	"github.com/mkarulina/loyalty-system-service.git/internal/authentication"
	"net/http"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("session_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("user not authorized"))
			return
		}

		auth := authentication.New()

		if token != nil && len(token.Value) >= 16 {
			valid, err := auth.CheckTokenIsValid(token.Value)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if valid {
				next.ServeHTTP(w, r)
				return
			}

			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("user not authorized"))
			return
		}
	})
}
