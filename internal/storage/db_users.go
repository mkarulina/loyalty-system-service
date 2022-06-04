package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/viper"
	"time"
)

func (s *storage) AddUserInfoToTable(user User) error {
	dbAddress := viper.GetString("DATABASE_DSN")

	db, err := sql.Open("pgx", dbAddress)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		"CREATE TABLE IF NOT EXISTS users (token VARCHAR(255), login VARCHAR(255), password VARCHAR(255))",
	)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := db.ExecContext(
		ctx,
		"INSERT INTO users (token, login, password) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
		user.Token, user.Login, user.Password,
	)
	if err != nil {
		return err
	}

	if affected, _ := result.RowsAffected(); affected < 1 {
		return errors.New(pgerrcode.UniqueViolation)
	}
	return nil
}

func (s *storage) CheckUserData(user User) error {
	dbAddress := viper.GetString("DATABASE_DSN")

	db, err := sql.Open("pgx", dbAddress)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := db.ExecContext(ctx, "SELECT * FROM users WHERE login = $1 AND password = $2", user.Login, user.Password)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return errors.New("user not registered")
	}

	_, err = db.ExecContext(ctx, "UPDATE users SET token = $1 WHERE login = $2 AND password = $3", user.Token, user.Login, user.Password)
	return nil
}

func (s *storage) CheckTokenIsValid(token string) (bool, error) {
	dbAddress := viper.GetString("DATABASE_DSN")

	db, err := sql.Open("pgx", dbAddress)
	if err != nil {
		return false, err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := db.ExecContext(ctx, "SELECT * FROM users WHERE token = $1", token)
	if err != nil {
		return false, err
	}

	if affected, _ := result.RowsAffected(); affected == 0 {
		return false, nil
	}

	return true, nil
}
