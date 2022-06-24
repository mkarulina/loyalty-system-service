package middleware

import (
	"net/http"
)

func (m *middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("session_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("user not authorized"))
			return
		}

		if token != nil && len(token.Value) >= 16 {
			valid, err := m.auth.CheckTokenIsValid(token.Value)
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
