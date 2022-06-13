package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/spf13/viper"
	"sync"
	"time"
)

func (s *storage) AddWithdrawnHistory(user string, order string, sum float32) error {
	dbAddress := viper.GetString("DATABASE_URI")
	var wg sync.WaitGroup
	var errToReturn error

	db, err := sql.Open("pgx", dbAddress)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doIncrement := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		defer wg.Done()

		_, err = db.ExecContext(
			ctx,
			"CREATE TABLE IF NOT EXISTS withdrawals_history ("+
				"user_id VARCHAR(255), "+
				"order_number VARCHAR(255), "+
				"sum FLOAT DEFAULT 0, "+
				"processed_at TIMESTAMP)",
		)
		if err != nil {
			errToReturn = err
			return
		}

		tx, err := db.Begin()
		if err != nil {
			errToReturn = err
			return
		}
		defer tx.Rollback()

		_, err = db.ExecContext(
			ctx,
			"INSERT INTO withdrawals_history (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4)",
			user, order, sum, time.Now().Format(time.RFC3339),
		)
		if err != nil {
			errToReturn = err
			return
		}
	}

	wg.Add(1)
	go doIncrement()
	wg.Wait()

	return errToReturn
}

func (s *storage) GetWithdrawalsHistory(token string) ([]Withdrawn, error) {
	dbAddress := viper.GetString("DATABASE_URI")
	var history []Withdrawn

	db, err := sql.Open("pgx", dbAddress)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	login, err := s.GetUserLoginByToken(token, db)
	if err != nil {
		return nil, err
	}
	if login == "" {
		return nil, errors.New("user not found")
	}

	result, err := db.QueryContext(ctx, "SELECT * FROM withdrawals_history WHERE user_id = $1", login)
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
