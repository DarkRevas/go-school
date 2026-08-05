package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// 5.2.1: Chain helper
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// 5.2.2: Request logging middleware
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(logger *slog.Logger) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			h.ServeHTTP(rec, r)
			logger.Info("request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// 5.2.3: Bearer token parser
var (
	ErrEmptyHeader   = errors.New("empty authorization header")
	ErrInvalidFormat = errors.New("invalid format, expected 'Bearer <token>'")
	ErrEmptyToken    = errors.New("empty token")
)

func extractBearerToken(header string) (string, error) {
	if header == "" {
		return "", ErrEmptyHeader
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrInvalidFormat
	}
	if parts[1] == "" {
		return "", ErrEmptyToken
	}
	return parts[1], nil
}

// 5.2.4: Auth middleware
type Claims struct {
	UserID string
}

type contextKey struct{}

var claimsKey contextKey

type TokenVerifier interface {
	Verify(token string) (Claims, error)
}

func AuthMiddleware(verifier TokenVerifier, logger *slog.Logger) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractBearerToken(r.Header.Get("Authorization"))
			if err != nil {
				logger.Warn("auth failed", "error", err)
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			claims, err := verifier.Verify(token)
			if err != nil {
				logger.Warn("token invalid", "error", err)
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// 5.2.5: JWT/Paseto adapters (mock implementations)
type JWTVerifier struct{}

func (JWTVerifier) Verify(token string) (Claims, error) {
	if token == "jwt-valid" {
		return Claims{UserID: "user-jwt"}, nil
	}
	return Claims{}, errors.New("invalid JWT token")
}

type PasetoVerifier struct{}

func (PasetoVerifier) Verify(token string) (Claims, error) {
	if token == "paseto-valid" {
		return Claims{UserID: "user-paseto"}, nil
	}
	return Claims{}, errors.New("invalid Paseto token")
}

// 5.2.6: Integration - API handlers
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(claimsKey).(Claims)
	if !ok {
		writeError(w, http.StatusInternalServerError, "claims not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"user_id":"` + claims.UserID + `"}`))
}

func main() {
	logger := slog.Default()

	mux := http.NewServeMux()

	// Public endpoint
	mux.HandleFunc("GET /health", healthHandler)

	// Protected endpoint with middleware chain
	protectedHandler := Chain(
		http.HandlerFunc(meHandler),
		AuthMiddleware(JWTVerifier{}, logger), 
	)
	mux.Handle("GET /api/v1/me", protectedHandler)

	logger.Info("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logger.Error("server failed", "error", err)
	}
}

// Задача 5.2.1. Chain helper

// Реализуй type Middleware func(http.Handler) http.Handler и функцию Chain.

// Проверь порядок исполнения на двух простых middleware, которые пишут логи "before/after".

// Задача 5.2.2. Request logging

// Сделай LoggingMiddleware на slog, который пишет:





// метод;



// путь;



// статус;



// длительность.

// Подключи middleware к GET /health.

// Задача 5.2.3. Bearer parser

// Сделай функцию extractBearerToken(header string) (string, error):





// корректно обрабатывай пустой заголовок;



// не принимай формат без префикса Bearer ;



// не принимай пустой токен.

// Задача 5.2.4. Auth middleware

// Определи интерфейс:

// type TokenVerifier interface {
// 	Verify(token string) (Claims, error)
// }

// Реализуй middleware, который:





// валидирует токен;



// кладет claims в context;



// на отказе возвращает 401.

// Задача 5.2.5. JWT/Paseto адаптер

// Сделай две реализации TokenVerifier:





// JWTVerifier (можно мок, без криптографии);



// PasetoVerifier (можно мок).

// Требование: HTTP-слой не должен меняться при переключении реализации.

// Задача 5.2.6. Интеграционная мини-задача

// Собери API:





// GET /health (публичный endpoint);



// GET /api/v1/me (только с валидным токеном).

// Требования:





// цепочка middleware для защищенного endpoint;



// структурированное логирование;



// единый формат 401 ответа;



// демонстрация, как подключить либо JWT, либо Paseto-верификатор без изменения handler'ов.