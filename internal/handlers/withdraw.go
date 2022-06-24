package handlers

import (
	"encoding/json"
	"github.com/jackc/pgerrcode"
	"io"
	"log"
	"net/http"
)

type withdrawReq struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

func (h *handler) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
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

	unmarshalBody := withdrawReq{}
	if err := json.Unmarshal(body, &unmarshalBody); err != nil {
		log.Println("can't unmarshal request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	balance, withdrawn, err := h.orderStg.GetUserBalanceAndWithdrawn(token.Value)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if balance-withdrawn < unmarshalBody.Sum {
		w.WriteHeader(http.StatusPaymentRequired)
		w.Write([]byte("insufficient funds to write off"))
		return
	}

	err = h.orderStg.AddOrderNumber(unmarshalBody.Order, token.Value)
	if err != nil {
		if err.Error() != pgerrcode.UniqueViolation && err.Error() != "duplicate" {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	err = h.orderStg.WithdrawUserPoints(token.Value, unmarshalBody.Order, unmarshalBody.Sum)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
