package handlers

import (
	"encoding/json"
	"github.com/mkarulina/loyalty-system-service.git/internal/authentication"
	"github.com/mkarulina/loyalty-system-service.git/internal/encryption"
	"io"
	"log"
	"net/http"
)

type loginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("session_token")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("can't read body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	unmarshalBody := loginReq{}
	if err := json.Unmarshal(body, &unmarshalBody); err != nil {
		log.Println("can't unmarshal request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e := encryption.New()
	encLogin := e.EncodeData(unmarshalBody.Login)
	encPassword := e.EncodeData(unmarshalBody.Password)

	err = h.auth.CheckUserData(authentication.User{
		Token:    token.Value,
		Login:    encLogin,
		Password: encPassword,
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
}
