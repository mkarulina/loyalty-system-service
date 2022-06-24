package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/mkarulina/loyalty-system-service.git/internal/authentication"
	"sync"
	"time"
)

type Withdrawn struct {
	UserID      string
	OrderNumber string
	Sum         float32
	ProcessedAt time.Time
}

type HistoryStorage interface {
	AddWithdrawnHistory(user string, order string, sum float32) error
	GetWithdrawalsHistory(token string) ([]Withdrawn, error)
}

type historyStorage struct {
	mu   sync.RWMutex
	db   *sql.DB
	auth authentication.Auth
}

func NewHistoryStorage() HistoryStorage {
	s := &historyStorage{
		mu:   sync.RWMutex{},
		db:   initDB(),
		auth: authentication.New(),
	}
	return s
}

func (s *historyStorage) AddWithdrawnHistory(user string, order string, sum float32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO withdrawals_history (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4)",
		user, order, sum, time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *historyStorage) GetWithdrawalsHistory(token string) ([]Withdrawn, error) {
	var history []Withdrawn

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	login, err := s.auth.GetUserLoginByToken(token, s.db)
	if err != nil {
		return nil, err
	}
	if login == "" {
		return nil, errors.New("user not found")
	}

	result, err := s.db.QueryContext(ctx, "SELECT * FROM withdrawals_history WHERE user_id = $1", login)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	for result.Next() {
		var w Withdrawn

		err = result.Scan(&w.UserID, &w.OrderNumber, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, err
		}
		history = append(history, w)
	}

	return history, nil
}
