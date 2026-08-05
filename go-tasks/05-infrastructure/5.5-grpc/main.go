package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	pb "github.com/my-project/token-service/pb"
)

// === 5.5.3: Token Repository Interface ===
type TokenRepository interface {
	Create(ctx context.Context, userID, scope string) (token string, expiresAt int64, err error)
	Validate(ctx context.Context, token string) (userID string, scope string, valid bool, err error)
	Revoke(ctx context.Context, token, reason string) error
}

// InMemoryTokenRepository - simple in-memory implementation
type InMemoryTokenRepository struct {
	mu     sync.RWMutex
	tokens map[string]*tokenData
}

type tokenData struct {
	userID    string
	scope     string
	expiresAt int64
	revoked   bool
}

func NewInMemoryTokenRepository() *InMemoryTokenRepository {
	return &InMemoryTokenRepository{
		tokens: make(map[string]*tokenData),
	}
}

func (r *InMemoryTokenRepository) Create(ctx context.Context, userID, scope string) (string, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	token := fmt.Sprintf("tok_%s_%d", userID, time.Now().UnixNano())
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	r.tokens[token] = &tokenData{
		userID:    userID,
		scope:     scope,
		expiresAt: expiresAt,
		revoked:   false,
	}

	return token, expiresAt, nil
}

func (r *InMemoryTokenRepository) Validate(ctx context.Context, token string) (string, string, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, exists := r.tokens[token]
	if !exists {
		return "", "", false, nil
	}
	if data.revoked {
		return "", "", false, nil
	}
	if data.expiresAt < time.Now().Unix() {
		return "", "", false, nil
	}

	return data.userID, data.scope, true, nil
}

func (r *InMemoryTokenRepository) Revoke(ctx context.Context, token, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, exists := r.tokens[token]
	if !exists {
		return errors.New("token not found")
	}

	data.revoked = true
	return nil
}

// === 5.5.3: Token Service Implementation ===
type TokenService struct {
	pb.UnimplementedTokenServiceServer
	repo   TokenRepository
	logger *slog.Logger
}

func NewTokenService(repo TokenRepository, logger *slog.Logger) *TokenService {
	return &TokenService{
		repo:   repo,
		logger: logger,
	}
}

func (s *TokenService) IssueToken(ctx context.Context, req *pb.IssueTokenRequest) (*pb.IssueTokenResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	token, expiresAt, err := s.repo.Create(ctx, req.UserId, req.Scope)
	if err != nil {
		s.logger.Error("failed to create token", "error", err, "user_id", req.UserId)
		return nil, status.Error(codes.Internal, "failed to issue token")
	}

	s.logger.Info("token issued", "user_id", req.UserId, "token", token[:10]+"...")
	return &pb.IssueTokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *TokenService) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	userID, scope, valid, err := s.repo.Validate(ctx, req.Token)
	if err != nil {
		s.logger.Error("failed to validate token", "error", err)
		return nil, status.Error(codes.Internal, "failed to validate token")
	}

	return &pb.ValidateTokenResponse{
		Valid:  valid,
		UserId: userID,
		Scope:  scope,
	}, nil
}

func (s *TokenService) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	err := s.repo.Revoke(ctx, req.Token, req.Reason)
	if err != nil {
		s.logger.Error("failed to revoke token", "error", err, "token", req.Token[:10]+"...")
		if err.Error() == "token not found" {
			return nil, status.Error(codes.NotFound, "token not found")
		}
		return nil, status.Error(codes.Internal, "failed to revoke token")
	}

	s.logger.Info("token revoked", "token", req.Token[:10]+"...", "reason", req.Reason)
	return &pb.RevokeTokenResponse{Revoked: true}, nil
}

// === 5.5.5: Unary Interceptors ===

// Logging interceptor
func LoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		logger.Debug("gRPC request started", "method", info.FullMethod)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			logger.Warn("gRPC request failed",
				"method", info.FullMethod,
				"duration", duration,
				"code", st.Code(),
				"error", err.Error(),
			)
		} else {
			logger.Debug("gRPC request completed",
				"method", info.FullMethod,
				"duration", duration,
			)
		}

		return resp, err
	}
}

// Auth interceptor for service-to-service authentication
var serviceToken = "service-secret-token-123"

func AuthInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Skip auth for ValidateToken (can be called publicly)
		if strings.HasSuffix(info.FullMethod, "ValidateToken") {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token := strings.TrimPrefix(tokens[0], "Bearer ")
		if token != serviceToken {
			logger.Warn("invalid service token", "method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		return handler(ctx, req)
	}
}

// Metrics interceptor (simple counter)
type MetricsCollector struct {
	mu       sync.Mutex
	requests map[string]int
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		requests: make(map[string]int),
	}
}

func (m *MetricsCollector) Increment(method string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[method]++
}

func (m *MetricsCollector) GetCounts() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]int)
	for k, v := range m.requests {
		result[k] = v
	}
	return result
}

func MetricsInterceptor(collector *MetricsCollector) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		collector.Increment(info.FullMethod)
		return handler(ctx, req)
	}
}

// === 5.5.4: gRPC Client ===
type TokenClient struct {
	conn   *grpc.ClientConn
	client pb.TokenServiceClient
}

func NewTokenClient(addr string) (*TokenClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				// Add authorization header
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+serviceToken)
				return invoker(ctx, method, req, reply, cc, opts...)
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &TokenClient{
		conn:   conn,
		client: pb.NewTokenServiceClient(conn),
	}, nil
}

func (c *TokenClient) IssueToken(ctx context.Context, userID, scope string) (string, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.IssueToken(ctx, &pb.IssueTokenRequest{
		UserId: userID,
		Scope:  scope,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return "", 0, fmt.Errorf("invalid request: %w", err)
			case codes.Unauthenticated:
				return "", 0, fmt.Errorf("authentication failed: %w", err)
			default:
				return "", 0, fmt.Errorf("gRPC error: %w", err)
			}
		}
		return "", 0, err
	}

	return resp.Token, resp.ExpiresAt, nil
}

func (c *TokenClient) ValidateToken(ctx context.Context, token string) (bool, string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		return false, "", "", err
	}

	return resp.Valid, resp.UserId, resp.Scope, nil
}

func (c *TokenClient) RevokeToken(ctx context.Context, token, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.RevokeToken(ctx, &pb.RevokeTokenRequest{
		Token:  token,
		Reason: reason,
	})
	return err
}

func (c *TokenClient) Close() error {
	return c.conn.Close()
}

// === 5.5.6: HTTP Gateway (Auth API) ===
type AuthAPI struct {
	tokenClient *TokenClient
	logger      *slog.Logger
}

func NewAuthAPI(client *TokenClient, logger *slog.Logger) *AuthAPI {
	return &AuthAPI{
		tokenClient: client,
		logger:      logger,
	}
}

func (a *AuthAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v1/tokens":
		if r.Method == http.MethodPost {
			a.issueTokenHandler(w, r)
			return
		}
	case "/api/v1/tokens/validate":
		if r.Method == http.MethodGet {
			a.validateTokenHandler(w, r)
			return
		}
	case "/api/v1/tokens/revoke":
		if r.Method == http.MethodPost {
			a.revokeTokenHandler(w, r)
			return
		}
	case "/health":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}
	http.NotFound(w, r)
}

func (a *AuthAPI) issueTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Scope  string `json:"scope"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	token, expiresAt, err := a.tokenClient.IssueToken(r.Context(), req.UserID, req.Scope)
	if err != nil {
		a.logger.Error("issue token failed", "error", err)
		if strings.Contains(err.Error(), "invalid request") {
			http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp, _ := protojson.Marshal(&pb.IssueTokenResponse{Token: token, ExpiresAt: expiresAt})
	w.Write(resp)
}

func (a *AuthAPI) validateTokenHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, `{"error":"token is required"}`, http.StatusBadRequest)
		return
	}

	valid, userID, scope, err := a.tokenClient.ValidateToken(r.Context(), token)
	if err != nil {
		a.logger.Error("validate token failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp, _ := protojson.Marshal(&pb.ValidateTokenResponse{Valid: valid, UserId: userID, Scope: scope})
	w.Write(resp)
}

func (a *AuthAPI) revokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token  string `json:"token"`
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	err := a.tokenClient.RevokeToken(r.Context(), req.Token, req.Reason)
	if err != nil {
		a.logger.Error("revoke token failed", "error", err)
		http.Error(w, `{"error":"token not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"revoked":true}`))
}

// === Main ===
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}).WithAttrs([]slog.Attr{
		slog.String("service", "token-service"),
	}))

	// Initialize repository and gRPC service
	repo := NewInMemoryTokenRepository()
	tokenService := NewTokenService(repo, logger)

	// Initialize metrics
	metrics := NewMetricsCollector()

	// Start gRPC server
	grpcAddr := ":50051"
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			chainUnaryInterceptors(
				LoggingInterceptor(logger),
				AuthInterceptor(logger),
				MetricsInterceptor(metrics),
			),
		),
	)
	pb.RegisterTokenServiceServer(grpcServer, tokenService)

	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("failed to create gRPC listener", "error", err)
		os.Exit(1)
	}

	go func() {
		logger.Info("gRPC server starting", "address", grpcAddr)
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Error("gRPC server failed", "error", err)
		}
	}()

	// Give gRPC server time to start
	time.Sleep(100 * time.Millisecond)

	// Create gRPC client
	client, err := NewTokenClient("localhost:50051")
	if err != nil {
		logger.Error("failed to create token client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	// Start HTTP gateway
	httpAddr := ":8080"
	authAPI := NewAuthAPI(client, logger)

	logger.Info("HTTP gateway starting", "address", httpAddr)
	if err := http.ListenAndServe(httpAddr, authAPI); err != nil && err != http.ErrServerClosed {
		logger.Error("HTTP server failed", "error", err)
	}
}

func chainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var chain grpc.UnaryHandler = handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			currentHandler := chain
			chain = func(ctx context.Context, req any) (any, error) {
				return interceptor(ctx, req, info, currentHandler)
			}
		}
		return chain(ctx, req)
	}
}
