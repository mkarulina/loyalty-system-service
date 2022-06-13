package handlers

import (
	"encoding/json"
	"github.com/jackc/pgerrcode"
	"io"
	"log"
	"net/http"
)

func (h *handler) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	type reqData struct {
		Order string  `json:"order"`
		Sum   float32 `json:"sum"`
	}

	token, err := r.Cookie("session_token")
	if err != nil || token == nil {
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

	unmarshalBody := reqData{}
	if err := json.Unmarshal(body, &unmarshalBody); err != nil {
		log.Println("can't unmarshal request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	balance, withdrawn, err := h.stg.GetUserBalanceAndWithdrawn(token.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if balance-withdrawn < unmarshalBody.Sum {
		w.WriteHeader(http.StatusPaymentRequired)
		w.Write([]byte("insufficient funds to write off"))
		return
	}

	err = h.stg.AddOrderNumber(unmarshalBody.Order, token.Value)
	if err != nil {
		if err.Error() != pgerrcode.UniqueViolation && err.Error() != "duplicate" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}

	err = h.stg.WithdrawUserPoints(token.Value, unmarshalBody.Order, unmarshalBody.Sum)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}
