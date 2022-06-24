package handlers

import (
	"encoding/json"
	"github.com/jackc/pgerrcode"
	"github.com/mkarulina/loyalty-system-service.git/internal/authentication"
	"github.com/mkarulina/loyalty-system-service.git/internal/encryption"
	"io"
	"log"
	"net/http"
)

type registerReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
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

	unmarshalBody := registerReq{}
	if err := json.Unmarshal(body, &unmarshalBody); err != nil {
		log.Println("can't unmarshal request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e := encryption.New()
	encLogin := e.EncodeData(unmarshalBody.Login)
	encPassword := e.EncodeData(unmarshalBody.Password)

	err = h.auth.AddUserInfoToTable(authentication.User{
		Token:    token.Value,
		Login:    encLogin,
		Password: encPassword,
	})
	if err != nil {
		if err.Error() == pgerrcode.UniqueViolation {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("username already exists"))
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
