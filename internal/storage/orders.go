package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/mkarulina/loyalty-system-service.git/internal/authentication"
	"log"
	"sync"
	"time"
)

type Order struct {
	UserID     string
	Number     string
	Status     string
	Accrual    float32
	Withdrawn  float32
	UploadedAt time.Time
}

type OrderStorage interface {
	AddOrderNumber(order string, token string) error
	GetUserOrders(token string) ([]Order, error)
	GetUserBalanceAndWithdrawn(token string) (float32, float32, error)
	WithdrawUserPoints(token string, order string, sum float32) error
	GetUnprocessedOrders() ([]string, error)
	UpdateOrdersStatus(orders []Order) error
}

type orderStorage struct {
	mu         sync.RWMutex
	db         *sql.DB
	auth       authentication.Auth
	historyStg HistoryStorage
}

func NewOrderStorage() OrderStorage {
	s := &orderStorage{
		mu:         sync.RWMutex{},
		db:         initDB(),
		auth:       authentication.New(),
		historyStg: NewHistoryStorage(),
	}
	return s
}

func (s *orderStorage) AddOrderNumber(order string, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	login, err := s.auth.GetUserLoginByToken(token, s.db)
	if err != nil {
		return err
	}
	if login == "" {
		return errors.New("user not found")
	}

	existingOrder, err := s.db.ExecContext(ctx, "SELECT * FROM orders WHERE number = $1", order)
	if err != nil {
		return err
	}

	if affected, _ := existingOrder.RowsAffected(); affected > 0 {

		orderByCurrentUser, err := s.db.ExecContext(ctx, "SELECT * FROM orders WHERE user_id = $2 AND number = $1", order, login)
		if err != nil {
			return err
		}

		if affected, _ = orderByCurrentUser.RowsAffected(); affected > 0 {
			return errors.New("duplicate")
		}
		return errors.New(pgerrcode.UniqueViolation)
	}

	_, err = s.db.ExecContext(
		ctx,
		"INSERT INTO orders (user_id, number, uploaded_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
		login, order, time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *orderStorage) GetUserOrders(token string) ([]Order, error) {
	var orders []Order

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	login, err := s.auth.GetUserLoginByToken(token, s.db)
	if err != nil || login == "" {
		return nil, err
	}

	ordersRows, err := s.db.QueryContext(ctx, "SELECT * FROM orders WHERE user_id = $1 ORDER BY uploaded_at ASC", login)
	if err != nil {
		return nil, err
	}
	if ordersRows.Err() != nil {
		return nil, ordersRows.Err()
	}

	for ordersRows.Next() {
		var o Order

		err = ordersRows.Scan(&o.UserID, &o.Number, &o.Status, &o.Accrual, &o.Withdrawn, &o.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func (s *orderStorage) GetUserBalanceAndWithdrawn(token string) (float32, float32, error) {
	var withdrawn float32
	var accrual float32

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	login, err := s.auth.GetUserLoginByToken(token, s.db)
	if err != nil {
		return 0, 0, err
	}
	if login == "" {
		return 0, 0, errors.New("user not found")
	}

	result, err := s.db.QueryContext(ctx, "SELECT SUM(accrual), SUM(withdrawn) FROM orders WHERE user_id = $1", login)
	if err != nil {
		return 0, 0, err
	}
	if result.Err() != nil {
		return 0, 0, result.Err()
	}

	for result.Next() {
		err = result.Scan(&accrual, &withdrawn)
		if err != nil {
			return 0, 0, err
		}
	}

	balance := accrual - withdrawn

	return balance, withdrawn, nil
}

func (s *orderStorage) WithdrawUserPoints(token string, order string, sum float32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.mu.Lock()

	login, err := s.auth.GetUserLoginByToken(token, s.db)
	if err != nil {
		return err
	}
	if login == "" {
		return errors.New("user not found")
	}

	result, err := s.db.ExecContext(ctx, "UPDATE orders SET withdrawn = $1 WHERE user_id = $2 AND number = $3", sum, login, order)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return errors.New("order not found")
	}

	s.mu.Unlock()

	err = s.historyStg.AddWithdrawnHistory(login, order, sum)
	if err != nil {
		log.Println("can't add withdrawn history")
	}

	return nil
}

func (s *orderStorage) GetUnprocessedOrders() ([]string, error) {
	var orders []string

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ordersRows, err := s.db.QueryContext(ctx, "SELECT number FROM orders WHERE status != 'PROCESSED'")
	if err != nil {
		return nil, err
	}
	if ordersRows.Err() != nil {
		return nil, ordersRows.Err()
	}

	for ordersRows.Next() {
		var order string

		err = ordersRows.Scan(&order)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *orderStorage) UpdateOrdersStatus(orders []Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update, err := s.db.PrepareContext(ctx, "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3")
	if err != nil {
		return err
	}
	defer update.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, order := range orders {
		_, err = update.ExecContext(ctx, order.Status, order.Accrual, order.Number)
		if err != nil {
			return err
		}
	}

	return nil
}
