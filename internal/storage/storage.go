package storage

import (
	"database/sql"
	"sync"
	"time"
)

type Storage interface {
	GetUserLoginByToken(token string, db *sql.DB) (string, error)
	AddOrderNumber(order string, token string) error
	GetUserOrders(token string) ([]Order, error)
	GetUserBalanceAndWithdrawn(token string) (float32, float32, error)
	WithdrawUserPoints(token string, order string, sum float32) error
	AddWithdrawnHistory(user string, order string, sum float32) error
	GetWithdrawalsHistory(token string) ([]Withdrawn, error)
	GetAllOrders() ([]string, error)
	UpdateOrdersStatus(orders []Order) error
}

type storage struct {
	mu sync.RWMutex
}

type User struct {
	Token    string
	Login    string
	Password string
}

type Order struct {
	UserID     string
	Number     string
	Status     string
	Accrual    float32
	Withdrawn  float32
	UploadedAt time.Time
}

type Withdrawn struct {
	UserID      string
	OrderNumber string
	Sum         float32
	ProcessedAt time.Time
}

func New() Storage {
	s := &storage{
		mu: sync.RWMutex{},
	}
	return s
}
