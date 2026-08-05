package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type tokensBody struct {
	UserId string `json:"user_id"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, 200, `{"status": "ok"}`)
	})

	mux.HandleFunc("GET /api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, 200, `{"access_token": r.PathValue("id")}`)
	})

	mux.HandleFunc("POST /api/v1/tokens", func(w http.ResponseWriter, r *http.Request) {
		var body tokensBody
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			WriteError(w, 400, errors.New("error invalid JSON"))
			return
		}
		if body.UserId == "" {
			WriteError(w, 400, errors.New("user_id is empty"))
			return
		}
		WriteJSON(w, 200, `{"access_token":"` + body.UserId + `"}`)
	})

	// Handler для теста graceful shutdown (длинный запрос)
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		log.Println("/slow started")
		time.Sleep(7 * time.Second) // Запрос выполняется 7 секунд (больше 5с таймаута)
		log.Println("/slow completed")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("done"))
	})

	// http.ListenAndServe(":8080", mux)

	server := &http.Server{
		Handler:           mux,
		Addr:              ":8080",
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       10 * time.Second,
	}

	go func() {
		log.Println("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received")

	graceCtx, close := context.WithTimeout(context.Background(), 5*time.Second)
	defer close()

	if err := server.Shutdown(graceCtx); err != nil {
		log.Fatalf("Graceful shutdown failed: %v", err)
	}

	log.Println("Server stopped gracefully")
}
// Собери мини-API "Token Service":
// GET /health;
// POST /api/v1/tokens;
// GET /api/v1/tokens/{id} (верни мок с данными токена).
// Требования:
// единый helper для JSON-ответов;
// единый helper для JSON-ошибок;
// маршруты описаны в стиле Go 1.22+;
// сервер запускается с таймаутами и graceful shutdown.

// Добавь обработку SIGINT/SIGTERM:
// при сигнале запускай Shutdown с таймаутом;
// логируй этапы остановки.
// Проверь вручную: отправь запрос в момент завершения и убедись, что сервер корректно закрывается.
// Переведи запуск с http.ListenAndServe на http.Server и задай:

// ReadHeaderTimeout
// ReadTimeout,
// WriteTimeout,
// IdleTimeout.

// Поясни в комментариях кода, почему выбраны именно такие значения.

// Подними HTTP-сервер на :8080 и реализуй маршрут:

// GET /health -> 200 OK, тело ok.

// Требование: используй http.NewServeMux() и шаблон маршрута с методом.

// Добавь маршрут:

// GET /api/v1/users/{id}

// Ответ:

// JSON вида {"user_id":"<id>"}.

// Требование: достань id через r.PathValue("id").

// Сделай endpoint:
// POST /api/v1/tokens
// Вход:
// {"user_id":"u-123"}
// Выход:
// при валидном запросе 201 и {"access_token":"..."};
// при невалидном JSON/пустом user_id -> 400 и JSON ошибки.
