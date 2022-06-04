package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/spf13/viper"
	"time"
)

func (s *storage) AddOrderNumber(order string, token string) error {
	dbAddress := viper.GetString("DATABASE_DSN")
	var login string

	db, err := sql.Open("pgx", dbAddress)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		"CREATE TABLE IF NOT EXISTS orders (user_id VARCHAR(255), number VARCHAR(255), status VARCHAR(255) DEFAULT 'NEW', accrual INTEGER DEFAULT 0, uploaded_at TIMESTAMP)",
	)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	loginRow := db.QueryRowContext(ctx, "SELECT login FROM users WHERE token = $1", token)
	if loginRow == nil {
		return errors.New("user not found")
	}
	err = loginRow.Scan(&login)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	existinпOrder, err := db.ExecContext(ctx, "SELECT * FROM orders WHERE number = $1", order)
	if err != nil {
		return err
	}

	if affected, _ := existinпOrder.RowsAffected(); affected > 0 {

		orderByCurrentUser, err := db.ExecContext(ctx, "SELECT * FROM orders WHERE user_id = $2 AND number = $1", order, login)
		if err != nil {
			return err
		}

		if affected, _ = orderByCurrentUser.RowsAffected(); affected > 0 {
			return errors.New("the order has already been created by the current user")
		}
		return errors.New(pgerrcode.UniqueViolation)
	}

	_, err = db.ExecContext(
		ctx,
		"INSERT INTO orders (user_id, number, uploaded_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
		login, order, time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *storage) GetUsersOrders(token string) ([]Order, error) {
	dbAddress := viper.GetString("DATABASE_DSN")
	var login string
	var orders []Order

	db, err := sql.Open("pgx", dbAddress)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	loginRow := db.QueryRowContext(ctx, "SELECT login FROM users WHERE token = $1", token)
	if loginRow == nil {
		return nil, errors.New("user not found")
	}
	err = loginRow.Scan(&login)
	if err != nil {
		return nil, err
	}

	ordersRows, err := db.QueryContext(ctx, "SELECT * FROM orders WHERE user_id = $1", login)
	if err != nil {
		return nil, err
	}

	for ordersRows.Next() {
		var o Order

		err = ordersRows.Scan(&o.UserId, &o.Number, &o.Status, &o.Accrual, &o.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}
