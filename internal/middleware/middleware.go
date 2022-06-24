package middleware

import (
	"github.com/mkarulina/loyalty-system-service.git/internal/authentication"
	"github.com/mkarulina/loyalty-system-service.git/internal/encryption"
	"net/http"
)

type Middleware interface {
	Auth(next http.Handler) http.Handler
	TokenHandle(next http.Handler) http.Handler
	GzipHandle(next http.Handler) http.Handler
}

type middleware struct {
	auth authentication.Auth
	enc  encryption.Encryptor
}

func NewMiddleware() Middleware {
	return &middleware{
		auth: authentication.New(),
		enc:  encryption.New(),
	}
}
