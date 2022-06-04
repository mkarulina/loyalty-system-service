package handlers

import (
	"encoding/json"
	"github.com/mkarulina/loyalty-system-service.git/internal/encryption"
	"github.com/mkarulina/loyalty-system-service.git/internal/storage"
	"io"
	"log"
	"net/http"
)

func (h *handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	type regData struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	token, err := r.Cookie("session_token")
	if err != nil {
		log.Println(err)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("can't read body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	unmarshalBody := regData{}
	if err := json.Unmarshal(body, &unmarshalBody); err != nil {
		log.Println("can't unmarshal request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e := encryption.New()
	encLogin := e.EncodeData(unmarshalBody.Login)
	encPassword := e.EncodeData(unmarshalBody.Password)

	err = h.stg.CheckUserData(storage.User{
		Token:    token.Value,
		Login:    encLogin,
		Password: encPassword,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}

}
