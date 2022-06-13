package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/spf13/viper"
	"log"
	"sync"
	"time"
)

func (s *storage) AddOrderNumber(order string, token string) error {
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
			"CREATE TABLE IF NOT EXISTS orders ("+
				"user_id VARCHAR(255), "+
				"number VARCHAR(255) UNIQUE, "+
				"status VARCHAR(255) DEFAULT 'NEW', "+
				"accrual FLOAT DEFAULT 0, "+
				"withdrawn FLOAT DEFAULT 0, "+
				"uploaded_at TIMESTAMP)",
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

		login, err := s.GetUserLoginByToken(token, db)
		if err != nil {
			errToReturn = err
			return
		}
		if login == "" {
			errToReturn = errors.New("user not found")
			return
		}

		existinпOrder, err := db.ExecContext(ctx, "SELECT * FROM orders WHERE number = $1", order)
		if err != nil {
			errToReturn = err
			return
		}

		if affected, _ := existinпOrder.RowsAffected(); affected > 0 {

			orderByCurrentUser, err := db.ExecContext(ctx, "SELECT * FROM orders WHERE user_id = $2 AND number = $1", order, login)
			if err != nil {
				errToReturn = err
			}

			if affected, _ = orderByCurrentUser.RowsAffected(); affected > 0 {
				errToReturn = errors.New("duplicate")
				return
			}
			errToReturn = errors.New(pgerrcode.UniqueViolation)
			return
		}

		_, err = db.ExecContext(
			ctx,
			"INSERT INTO orders (user_id, number, uploaded_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
			login, order, time.Now().Format(time.RFC3339),
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

func (s *storage) GetUserOrders(token string) ([]Order, error) {
	dbAddress := viper.GetString("DATABASE_URI")
	var orders []Order

	db, err := sql.Open("pgx", dbAddress)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	login, err := s.GetUserLoginByToken(token, db)
	if err != nil || login == "" {
		return nil, err
	}

	ordersRows, err := db.QueryContext(ctx, "SELECT * FROM orders WHERE user_id = $1 ORDER BY uploaded_at ASC", login)
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

func (s *storage) GetUserBalanceAndWithdrawn(token string) (float32, float32, error) {
	dbAddress := viper.GetString("DATABASE_URI")
	var withdrawn float32
	var accrual float32

	db, err := sql.Open("pgx", dbAddress)
	if err != nil {
		return 0, 0, err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	login, err := s.GetUserLoginByToken(token, db)
	if err != nil {
		return 0, 0, err
	}
	if login == "" {
		return 0, 0, errors.New("user not found")
	}

	result, err := db.QueryContext(ctx, "SELECT SUM(accrual), SUM(withdrawn) FROM orders WHERE user_id = $1", login)
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

func (s *storage) WithdrawUserPoints(token string, order string, sum float32) error {
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
		defer wg.Done()

		login, err := s.GetUserLoginByToken(token, db)
		if err != nil {
			errToReturn = err
			return
		}
		if login == "" {
			errToReturn = errors.New("user not found")
			return
		}

		result, err := db.ExecContext(ctx, "UPDATE orders SET withdrawn = $1 WHERE user_id = $2 AND number = $3", sum, login, order)
		if err != nil {
			errToReturn = err
			return
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			errToReturn = errors.New("order not found")
			return
		}

		s.mu.Unlock()

		err = s.AddWithdrawnHistory(login, order, sum)
		if err != nil {
			log.Println("can't add withdrawn history")
		}
	}

	wg.Add(1)
	go doIncrement()
	wg.Wait()

	return errToReturn
}

func (s *storage) GetAllOrders() ([]string, error) {
	dbAddress := viper.GetString("DATABASE_URI")
	var orders []string

	db, err := sql.Open("pgx", dbAddress)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ordersRows, err := db.QueryContext(ctx, "SELECT number FROM orders")
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

func (s *storage) UpdateOrdersStatus(orders []Order) error {
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

	update, err := db.PrepareContext(ctx, "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3")
	if err != nil {
		return err
	}
	defer update.Close()

	doIncrement := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		defer wg.Done()

		for _, order := range orders {
			_, err := update.ExecContext(ctx, order.Status, order.Accrual, order.Number)
			if err != nil {
				errToReturn = err
			}
		}
	}

	wg.Add(1)
	go doIncrement()
	wg.Wait()

	return errToReturn
}
