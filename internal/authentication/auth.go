package authentication

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/spf13/viper"
	"sync"
	"time"
)

type User struct {
	Token    string
	Login    string
	Password string
}

type Auth interface {
	AddUserInfoToTable(user User) error
	CheckUserData(user User) error
	CheckTokenIsValid(token string) (bool, error)
	GetUserLoginByToken(token string, db *sql.DB) (string, error)
}

type auth struct {
	mu sync.RWMutex
}

func New() Auth {
	a := &auth{
		mu: sync.RWMutex{},
	}
	return a
}

func (a *auth) AddUserInfoToTable(user User) error {
	dbAddress := viper.GetString("DATABASE_URI")

	db, err := sql.Open("pgx", dbAddress)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.mu.Lock()
	defer a.mu.Unlock()

	result, err := db.ExecContext(
		ctx,
		"INSERT INTO users (token, login, password) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
		user.Token, user.Login, user.Password,
	)
	if err != nil {
		return err
	}

	if affected, _ := result.RowsAffected(); affected == 0 {
		return errors.New(pgerrcode.UniqueViolation)
	}

	return nil
}

func (a *auth) CheckUserData(user User) error {
	dbAddress := viper.GetString("DATABASE_URI")

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

	a.mu.Lock()
	defer a.mu.Unlock()

	_, err = db.ExecContext(ctx, "UPDATE users SET token = $1 WHERE login = $2 AND password = $3", user.Token, user.Login, user.Password)
	if err != nil {
		return err
	}

	return nil
}

func (a *auth) CheckTokenIsValid(token string) (bool, error) {
	dbAddress := viper.GetString("DATABASE_URI")

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

func (a *auth) GetUserLoginByToken(token string, db *sql.DB) (string, error) {
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
