package storage

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/jackc/pgx/v4/stdlib"
	"time"
)

func (s *storage) GetUserLoginByToken(token string, db *sql.DB) (string, error) {
	var login string

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	loginRow := db.QueryRowContext(ctx, "SELECT login FROM users WHERE token = $1", token)
	if loginRow == nil {
		return "", errors.New("user not found")
	}
	err := loginRow.Scan(&login)
	if err != nil {
		return "", err
	}
	return login, nil
}
