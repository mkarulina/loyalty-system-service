package main

import (
	"compress/gzip"
	"encoding/hex"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mkarulina/loyalty-system-service.git/config"
	"github.com/mkarulina/loyalty-system-service.git/internal/encryption"
	"github.com/mkarulina/loyalty-system-service.git/internal/handlers"
	"github.com/mkarulina/loyalty-system-service.git/internal/storage"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func main() {
	_, err := config.LoadConfig("config")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	h := handlers.NewHandler(storage.New())

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/user", func(r chi.Router) {

		r.Route("/register", func(r chi.Router) {
			r.Use(tokenHandle)
			r.Post("/", h.RegisterHandler)
		})

		r.Route("/login", func(r chi.Router) {
			r.Use(tokenHandle)
			r.Post("/", h.LoginHandler)
		})

		r.Route("/", func(r chi.Router) {
			r.Use(auth)
			r.Use(gzipHandle)
			r.Post("/orders", h.SendOrderHandler)                         //загрузка пользователем номера заказа для расчёта
			r.Get("/orders", h.GetOrderHandler)                           //получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
			r.Get("/balance", h.GetBalanceHandler)                        //получение текущего баланса счёта баллов лояльности пользователя
			r.Post("/balance/withdraw", h.WithdrawHandler)                //запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
			r.Get("/balance/withdrawals", h.GetWithdrawalsHistoryHandler) //получение информации о выводе средств с накопительного счёта пользователем
		})
	})

	address := viper.GetString("SERVER_ADDRESS")

	if err := http.ListenAndServe(address, r); err != nil {
		log.Fatal(err)
	}
}

func (gw gzipWriter) Write(b []byte) (int, error) {
	return gw.Writer.Write(b)
}

func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.Header.Get(`Content-Encoding`), `gzip`) {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gz
			next.ServeHTTP(w, r)
			return
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stg := storage.New()

		token, err := r.Cookie("session_token")
		if err != nil {
			w.Write([]byte("user not authorized"))
			return
		}

		if token != nil && len(token.Value) >= 16 {
			valid, err := stg.CheckTokenIsValid(token.Value)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if valid {
				next.ServeHTTP(w, r)
				return
			}

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("user not authorized"))
			return
		}
	})
}

func tokenHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("session_token")
		if token != nil && len(token.Value) >= 16 {
			next.ServeHTTP(w, r)
			return
		}
		if err != nil {
			log.Println(err)
		}

		e := encryption.New()

		random, err := e.GenerateRandom(16)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newUser := hex.EncodeToString(random)

		newToken, err := e.EncryptData([]byte(newUser))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newCookie := &http.Cookie{
			Name:    "session_token",
			Value:   newToken,
			Expires: time.Now().Add(3 * time.Hour),
			Secure:  false,
		}

		http.SetCookie(w, newCookie)
		r.AddCookie(newCookie)
		next.ServeHTTP(w, r)
	})
}
