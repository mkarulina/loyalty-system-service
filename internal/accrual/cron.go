package accrual

import (
	"encoding/json"
	"github.com/mkarulina/loyalty-system-service.git/internal/storage"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
)

func StartCron() {
	c := cron.New()
	s := storage.NewOrderStorage()

	c.AddFunc("@every 10s", func() {
		GetOrdersStatus(s)
	})

	c.Start()
}

func GetOrdersStatus(s storage.OrderStorage) {
	type accrualResp struct {
		Order   string  `json:"order"`
		Status  string  `json:"status"`
		Accrual float32 `json:"accrual"`
	}
	var ordersStatus []storage.Order

	accrualAddress := viper.GetString("ACCRUAL_SYSTEM_ADDRESS")

	orders, err := s.GetUnprocessedOrders()
	if err != nil {
		log.Println(err)
	}

	for _, order := range orders {
		response, err := http.Get(accrualAddress + "/api/orders/" + order)
		if err != nil {
			log.Println(err)
		}

		switch response.StatusCode {
		case http.StatusInternalServerError:
			log.Println("accrual response code: ", http.StatusInternalServerError)
		case http.StatusTooManyRequests:
			log.Println("accrual response code: ", http.StatusTooManyRequests)
		case http.StatusNoContent:
			log.Println("accrual response code: ", http.StatusNoContent)
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Println(err)
		}
		response.Body.Close()

		data := accrualResp{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Println(err)
		}

		orderStatus := storage.Order{
			Number:  data.Order,
			Status:  data.Status,
			Accrual: data.Accrual,
		}

		ordersStatus = append(ordersStatus, orderStatus)
	}

	err = s.UpdateOrdersStatus(ordersStatus)
	if err != nil {
		log.Println("order update error: ", err)
	}
}
