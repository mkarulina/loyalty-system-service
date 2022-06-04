package storage

import (
	"sync"
	"time"
)

type Storage interface {
	AddUserInfoToTable(user User) error
	CheckUserData(user User) error
	CheckTokenIsValid(token string) (bool, error)
	AddOrderNumber(order string, token string) error
	GetUsersOrders(token string) ([]Order, error)
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
	UserId     string
	Number     string
	Status     string
	Accrual    int
	UploadedAt time.Time
}

func New() Storage {
	s := &storage{
		mu: sync.RWMutex{},
	}
	return s
}
