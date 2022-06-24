package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type withdrawalsHistoryResp struct {
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (h *handler) GetWithdrawalsHistoryHandler(w http.ResponseWriter, r *http.Request) {
	var resp []withdrawalsHistoryResp

	token, err := r.Cookie("session_token")
	if err != nil || token == nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	withdrawals, err := h.historyStg.GetWithdrawalsHistory(token.Value)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for _, w := range withdrawals {
		resp = append(resp, withdrawalsHistoryResp{
			Order:       w.OrderNumber,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt,
		})
	}

	marshalResp, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(marshalResp)
}
