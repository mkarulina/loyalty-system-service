package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func (h *handler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	type balanceResp struct {
		Current   float32 `json:"current"`
		Withdrawn float32 `json:"withdrawn"`
	}

	token, err := r.Cookie("session_token")
	if err != nil || token == nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	balance, withdrawn, err := h.stg.GetUserBalanceAndWithdrawn(token.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	resp := balanceResp{
		Current:   balance,
		Withdrawn: withdrawn,
	}

	marshalResp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(marshalResp)
}
