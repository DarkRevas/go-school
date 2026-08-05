package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Domain errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrAccountNotFound   = errors.New("account not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrTokenNotFound     = errors.New("token not found")
	ErrTokenRevoked      = errors.New("token revoked")
	ErrTokenExpired      = errors.New("token expired")
)

// User represents a user in the system
type User struct {
	ID        int64
	Email     string
	CreatedAt time.Time
}

// Account represents a user account with balance
type Account struct {
	ID        int64
	UserID    int64
	Balance   int64
	CreatedAt time.Time
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}

// UserRepository handles user data operations
type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, email string) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (email) VALUES ($1) RETURNING id, email, created_at
	`, email).Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, created_at FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// AccountRepository handles account data operations
type AccountRepository struct {
	pool *pgxpool.Pool
}

func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool}
}

func (r *AccountRepository) Create(ctx context.Context, userID int64, balance int64) (Account, error) {
	var a Account
	err := r.pool.QueryRow(ctx, `
		INSERT INTO accounts (user_id, balance) VALUES ($1, $2) RETURNING id, user_id, balance, created_at
	`, userID, balance).Scan(&a.ID, &a.UserID, &a.Balance, &a.CreatedAt)
	if err != nil {
		return Account{}, fmt.Errorf("create account: %w", err)
	}
	return a, nil
}

func (r *AccountRepository) GetByUserID(ctx context.Context, userID int64) (Account, error) {
	var a Account
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, balance, created_at FROM accounts WHERE user_id = $1
	`, userID).Scan(&a.ID, &a.UserID, &a.Balance, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrAccountNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("get account by user id: %w", err)
	}
	return a, nil
}

// Transfer moves funds between accounts in a transaction
func (r *AccountRepository) Transfer(ctx context.Context, fromID, toID, amount int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check and deduct from source account
	var fromBalance int64
	err = tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id = $1 FOR UPDATE`, fromID).Scan(&fromBalance)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("source account: %w", ErrAccountNotFound)
	}
	if err != nil {
		return fmt.Errorf("get source balance: %w", err)
	}
	if fromBalance < amount {
		return ErrInsufficientFunds
	}

	_, err = tx.Exec(ctx, `UPDATE accounts SET balance = balance - $1 WHERE id = $2`, amount, fromID)
	if err != nil {
		return fmt.Errorf("debit source: %w", err)
	}

	// Credit destination account
	_, err = tx.Exec(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amount, toID)
	if err != nil {
		return fmt.Errorf("credit destination: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// TokenRepository handles refresh token operations
type TokenRepository struct {
	pool *pgxpool.Pool
}

func NewTokenRepository(pool *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{pool: pool}
}

func (r *TokenRepository) Create(ctx context.Context, userID int64, token string, expiresAt time.Time) (RefreshToken, error) {
	var rt RefreshToken
	err := r.pool.QueryRow(ctx, `
		INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)
		RETURNING id, user_id, token, expires_at, revoked, created_at
	`, userID, token, expiresAt).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("create refresh token: %w", err)
	}
	return rt, nil
}

func (r *TokenRepository) GetActive(ctx context.Context, token string) (RefreshToken, error) {
	var rt RefreshToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token, expires_at, revoked, created_at
		FROM refresh_tokens WHERE token = $1 AND revoked = FALSE AND expires_at > NOW()
	`, token).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return RefreshToken{}, ErrTokenNotFound
	}
	if err != nil {
		return RefreshToken{}, fmt.Errorf("get active token: %w", err)
	}
	return rt, nil
}

func (r *TokenRepository) Revoke(ctx context.Context, token string) error {
	result, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1`, token)
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrTokenNotFound
	}
	return nil
}

// TokenService provides business logic for token operations
type TokenService struct {
	userRepo  *UserRepository
	tokenRepo *TokenRepository
}

func NewTokenService(userRepo *UserRepository, tokenRepo *TokenRepository) *TokenService {
	return &TokenService{userRepo: userRepo, tokenRepo: tokenRepo}
}

// LoginAndIssueToken creates a user and issues a refresh token in a transaction
func (s *TokenService) LoginAndIssueToken(ctx context.Context, email string) (User, RefreshToken, error) {
	tx, err := s.userRepo.pool.Begin(ctx)
	if err != nil {
		return User{}, RefreshToken{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create or get user
	var user User
	err = tx.QueryRow(ctx, `INSERT INTO users (email) VALUES ($1) ON CONFLICT (email) DO UPDATE SET email = $1 RETURNING id, email, created_at`, email).
		Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		return User{}, RefreshToken{}, fmt.Errorf("create/get user: %w", err)
	}

	// Create refresh token
	token := fmt.Sprintf("refresh_%d_%d", user.ID, time.Now().UnixNano())
	expiresAt := time.Now().Add(24 * time.Hour)
	var rt RefreshToken
	err = tx.QueryRow(ctx, `
		INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)
		RETURNING id, user_id, token, expires_at, revoked, created_at
	`, user.ID, token, expiresAt).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt)
	if err != nil {
		return User{}, RefreshToken{}, fmt.Errorf("create token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, RefreshToken{}, fmt.Errorf("commit transaction: %w", err)
	}
	return user, rt, nil
}

func initDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Ping to verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.Default()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/token_service?sslmode=disable"
	}

	pool, err := initDB(ctx, dsn)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("database connected")

	// Initialize repositories
	userRepo := NewUserRepository(pool)
	accountRepo := NewAccountRepository(pool)
	tokenRepo := NewTokenRepository(pool)
	tokenService := NewTokenService(userRepo, tokenRepo)

	user, token, err := tokenService.LoginAndIssueToken(ctx, "test@example.com")
	if err != nil {
		logger.Error("login failed", "error", err)
	} else {
		logger.Info("login successful", "user_id", user.ID, "token", token.Token)
	}

	err = accountRepo.Transfer(ctx, 1, 2, 100)
	if err != nil {
		logger.Error("transfer failed", "error", err)
	}

	logger.Info("waiting for shutdown signal")
	<-ctx.Done()
	logger.Info("shutting down gracefully")
}
