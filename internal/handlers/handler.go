package handlers

import (
	authentication "github.com/mkarulina/loyalty-system-service.git/internal/authentication"
	"github.com/mkarulina/loyalty-system-service.git/internal/storage"
	"net/http"
)

type Handler interface {
	RegisterHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
	SendOrderHandler(w http.ResponseWriter, r *http.Request)
	GetOrderHandler(w http.ResponseWriter, r *http.Request)
	GetBalanceHandler(w http.ResponseWriter, r *http.Request)
	WithdrawHandler(w http.ResponseWriter, r *http.Request)
	GetWithdrawalsHistoryHandler(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	stg  storage.Storage
	auth authentication.Auth
}

func NewHandler(s storage.Storage) Handler {
	h := &handler{
		stg:  s,
		auth: authentication.New(),
	}
	return h
}
