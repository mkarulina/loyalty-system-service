package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func (h *handler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	type orderResp struct {
		Number     string  `json:"number"`
		Status     string  `json:"status"`
		Accrual    float32 `json:"accrual,omitempty"`
		UploadedAt string  `json:"uploaded_at"`
	}
	var resp []orderResp

	token, err := r.Cookie("session_token")
	if err != nil || token == nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orders, err := h.stg.GetUserOrders(token.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for _, o := range orders {
		resp = append(resp, orderResp{
			Number:     o.Number,
			Status:     o.Status,
			Accrual:    o.Accrual,
			UploadedAt: o.UploadedAt.Format(time.RFC3339),
		})
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
