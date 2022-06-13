package authentication

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/mkarulina/loyalty-system-service.git/internal/storage"
	"github.com/spf13/viper"
	"sync"
	"time"
)

type Auth interface {
	AddUserInfoToTable(user storage.User) error
	CheckUserData(user storage.User) error
	CheckTokenIsValid(token string) (bool, error)
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

func (a *auth) AddUserInfoToTable(user storage.User) error {
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
		a.mu.Lock()
		defer a.mu.Unlock()
		defer wg.Done()

		_, err = db.ExecContext(
			ctx,
			"CREATE TABLE IF NOT EXISTS users (token VARCHAR(255), login VARCHAR(255) UNIQUE, password VARCHAR(255))",
		)
		if err != nil {
			errToReturn = err
			return
		}

		tx, err := db.Begin()
		if err != nil {
			errToReturn = err
		}
		defer tx.Rollback()

		result, err := db.ExecContext(
			ctx,
			"INSERT INTO users (token, login, password) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
			user.Token, user.Login, user.Password,
		)
		if err != nil {
			errToReturn = err
			return
		}

		if affected, _ := result.RowsAffected(); affected == 0 {
			errToReturn = errors.New(pgerrcode.UniqueViolation)
			return
		}
	}

	wg.Add(1)
	go doIncrement()
	wg.Wait()

	return errToReturn
}

func (a *auth) CheckUserData(user storage.User) error {
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

	result, err := db.ExecContext(ctx, "SELECT * FROM users WHERE login = $1 AND password = $2", user.Login, user.Password)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return errors.New("user not registered")
	}
	
	doIncrement := func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		defer wg.Done()

		_, err = db.ExecContext(ctx, "UPDATE users SET token = $1 WHERE login = $2 AND password = $3", user.Token, user.Login, user.Password)
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
