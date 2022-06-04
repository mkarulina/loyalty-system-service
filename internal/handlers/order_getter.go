package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *handler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	type orderResp struct {
		Number      string `json:"number"`
		Status      string `json:"status"`
		Accrual     int    `json:"accrual,omitempty"`
		Uploaded_at string `json:"uploaded_at"`
	}
	var resp []orderResp

	token, err := r.Cookie("session_token")
	if err != nil || token == nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	orders, err := h.stg.GetUsersOrders(token.Value)
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
			Number:      o.Number,
			Status:      o.Status,
			Accrual:     o.Accrual,
			Uploaded_at: o.UploadedAt.String(),
		})
	}

	marshalResp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(marshalResp)
}
