package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// 5.4.1: Structured logger with slog
var (
	serviceName = os.Getenv("SERVICE_NAME")
	env         = os.Getenv("ENV")
	version     = os.Getenv("VERSION")
)

func init() {
	if serviceName == "" {
		serviceName = "token-service"
	}
	if env == "" {
		env = "development"
	}
	if version == "" {
		version = "0.0.0"
	}
}

func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}).WithAttrs([]slog.Attr{
		slog.String("service", serviceName),
		slog.String("env", env),
		slog.String("version", version),
	}))
}

func LogError(logger *slog.Logger, msg string, err error, attrs ...any) {
	logger.Error(msg, append(attrs, "error", err.Error())...)
}

// 5.4.2: HTTP metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, httpRequestsInFlight)
}

// 5.4.3: Route normalization
var routePatterns = []*regexp.Regexp{
	regexp.MustCompile(`/api/v1/users/\d+`),
	regexp.MustCompile(`/api/v1/tokens/\w+`),
	regexp.MustCompile(`/api/v1/accounts/\d+`),
}

func normalizeRoute(path string) string {
	for _, pattern := range routePatterns {
		if pattern.MatchString(path) {
			matches := pattern.FindStringSubmatch(path)
			if len(matches) > 0 {
				return pattern.ReplaceAllString(path, pattern.String())
			}
		}
	}
	// Generic normalization for numeric IDs
	normalized := regexp.MustCompile(`/\d+`).ReplaceAllString(path, "/{id}")
	normalized = regexp.MustCompile(`/\w{8,}`).ReplaceAllString(normalized, "/{token}")
	return normalized
}

// 5.4.2 + 5.4.3: Metrics middleware
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		httpRequestsInFlight.Inc()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		route := normalizeRoute(r.URL.Path)
		status := strconv.Itoa(rec.status)

		httpRequestsTotal.WithLabelValues(r.Method, route, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
		httpRequestsInFlight.Dec()
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// 5.4.4: OpenTelemetry tracing
func NewTracerProvider() (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("create stdout exporter: %w", err)
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
			semconv.DeploymentEnvironment(env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// 5.4.5: Correlation - extract trace/span IDs for logging
func TraceContextToAttrs(ctx context.Context) []slog.Attr {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return nil
	}

	spanCtx := span.SpanContext()
	if !spanCtx.IsValid() {
		return nil
	}

	return []slog.Attr{
		slog.String("trace_id", spanCtx.TraceID().String()),
		slog.String("span_id", spanCtx.SpanID().String()),
	}
}

func LoggerWithTrace(ctx context.Context, logger *slog.Logger) *slog.Logger {
	attrs := TraceContextToAttrs(ctx)
	if len(attrs) == 0 {
		return logger
	}
	args := make([]any, 0, len(attrs)*2)
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value.Any())
	}
	return logger.With(args...)
}

// Tracing middleware
func TracingMiddleware(tracer trace.Tracer, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "HTTP "+r.Method+" "+r.URL.Path,
				trace.WithAttributes(
					semconv.HTTPMethod(r.Method),
					semconv.HTTPRoute(r.URL.Path),
					semconv.URLPath(r.URL.Path),
				),
			)
			defer span.End()

			r = r.WithContext(ctx)
			logger := LoggerWithTrace(ctx, logger)

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			span.SetAttributes(semconv.HTTPStatusCode(rec.status))
			if rec.status >= 400 {
				span.SetStatus(1, "HTTP error")
			}

			logger.Debug("request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
			)
		})
	}
}

// Token Service with observability
type TokenService struct {
	tracer trace.Tracer
	logger *slog.Logger
}

func NewTokenService(tracer trace.Tracer, logger *slog.Logger) *TokenService {
	return &TokenService{tracer: tracer, logger: logger}
}

func (s *TokenService) IssueToken(ctx context.Context, userID string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "IssueToken",
		trace.WithAttributes(attribute.String("user_id", userID)),
	)
	defer span.End()

	logger := LoggerWithTrace(ctx, s.logger)
	logger.Info("issuing token", slog.String("user_id", userID))

	// Simulate token generation
	if userID == "" {
		err := errors.New("user_id is empty")
		LogError(logger, "token issuance failed", err)
		span.RecordError(err)
		span.SetStatus(1, "token issuance failed")
		return "", err
	}

	token := fmt.Sprintf("token_%s_%d", userID, time.Now().UnixNano())
	logger.Info("token issued", slog.String("token", token))

	return token, nil
}

func (s *TokenService) ValidateToken(ctx context.Context, token string) error {
	ctx, span := s.tracer.Start(ctx, "ValidateToken",
		trace.WithAttributes(attribute.String("token", token)),
	)
	defer span.End()

	logger := LoggerWithTrace(ctx, s.logger)
	logger.Debug("validating token", slog.String("token", token))

	if !strings.HasPrefix(token, "token_") {
		err := errors.New("invalid token format")
		LogError(logger, "token validation failed", err)
		span.RecordError(err)
		return err
	}

	logger.Info("token validated")
	return nil
}

// HTTP handlers
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func metricsHandler() http.Handler {
	return promhttp.Handler()
}

func issueTokenHandler(svc *TokenService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := LoggerWithTrace(ctx, logger)

		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"user_id is required"}`))
			return
		}

		token, err := svc.IssueToken(ctx, userID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"` + err.Error() + `"}`))
			return
		}

		logger.Info("token issued via HTTP")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"token":"` + token + `"}`))
	}
}

func validateTokenHandler(svc *TokenService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := LoggerWithTrace(ctx, logger)

		token := r.URL.Query().Get("token")
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"token is required"}`))
			return
		}

		err := svc.ValidateToken(ctx, token)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"` + err.Error() + `"}`))
			return
		}

		logger.Info("token validated via HTTP")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"valid":true}`))
	}
}

func main() {
	logger := NewLogger()
	logger.Info("starting token service")

	tp, err := NewTracerProvider()
	if err != nil {
		LogError(logger, "failed to create tracer provider", err)
		os.Exit(1)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("failed to shutdown tracer", "error", err)
		}
	}()

	tracer := otel.Tracer("token-service")
	tokenService := NewTokenService(tracer, logger)

	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("GET /metrics", metricsHandler())

	// API endpoints with tracing
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("POST /api/v1/tokens", issueTokenHandler(tokenService, logger))
	apiMux.HandleFunc("GET /api/v1/tokens/validate", validateTokenHandler(tokenService, logger))

	// Apply middleware chain
	handler := http.Handler(apiMux)
	handler = TracingMiddleware(tracer, logger)(handler)
	handler = MetricsMiddleware(handler)

	mux.Handle("/api/v1/", handler)

	logger.Info("server starting on :8080",
		slog.String("address", ":8080"),
		slog.String("metrics", "http://localhost:8080/metrics"),
	)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		LogError(logger, "server failed", err)
		os.Exit(1)
	}
}
