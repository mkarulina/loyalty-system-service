package middleware

import (
	"encoding/hex"
	"log"
	"net/http"
	"time"
)

func (m *middleware) TokenHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("session_token")
		if token != nil && len(token.Value) >= 16 {
			next.ServeHTTP(w, r)
			return
		}
		if err != nil {
			log.Println(err)
		}

		random, err := m.enc.GenerateRandom(16)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newUser := hex.EncodeToString(random)

		newToken, err := m.enc.EncryptData([]byte(newUser))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newCookie := &http.Cookie{
			Name:    "session_token",
			Value:   newToken,
			Expires: time.Now().Add(3 * time.Hour),
			Secure:  false,
		}

		http.SetCookie(w, newCookie)
		r.AddCookie(newCookie)
		next.ServeHTTP(w, r)
	})
}
