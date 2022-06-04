package handlers

import (
	"github.com/jackc/pgerrcode"
	"io"
	"net/http"
	"strings"
)

func (h *handler) SendOrderHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get(`Content-Type`), `text/plain`) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request format"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := r.Cookie("session_token")
	if err != nil || token == nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	reqValue := string(body)

	//todo: проверка номера заказа

	//validOrder := govalidator.IsType(reqValue, "int")
	//
	//if !validOrder {
	//	w.WriteHeader(http.StatusUnprocessableEntity)
	//	w.Write([]byte("Проверьте формат номера заказа"))
	//	return
	//}

	err = h.stg.AddOrderNumber(reqValue, token.Value)
	if err != nil {
		if err.Error() == "the order has already been created by the current user" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(err.Error()))
			return
		}

		if err.Error() == pgerrcode.UniqueViolation {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("the order has already been created by the another user"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
