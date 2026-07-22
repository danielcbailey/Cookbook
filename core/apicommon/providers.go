package apicommon

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/google/uuid"
)

type providersHandler struct {
	next      http.Handler
	providers config.Providers
}

type providersCtxKey struct{}

type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	errorBody   []byte
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.wroteHeader = true
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	if rw.statusCode >= 400 {
		rw.errorBody = append(rw.errorBody, b...)
	}
	return rw.ResponseWriter.Write(b)
}

func WithLoggingProviders(providers config.Providers, next http.Handler) http.Handler {
	return &providersHandler{next: next, providers: providers}
}

func (h *providersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Checking IP rate limit
	rl := IPRateLimiter(h.providers.Cache(), r.RemoteAddr)
	if ok, err := rl.Add(r.Context(), 1); err != nil {
		// Since this is to do with rate limiting, a log shouldn't be emitted as a severe DDoS
		// attack could overwelm the cache and cause it to error. Instead, a distinctive error
		// is sent to the client.
		http.Error(w, "failed to check rate limit", http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	requestID := uuid.Must(uuid.NewV7()).String()
	providers := h.providers.WithLog(h.providers.Log().With(slog.String("xid", requestID)))

	ctx := context.WithValue(r.Context(), providersCtxKey{}, providers)
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	h.next.ServeHTTP(rw, r.WithContext(ctx))
	attrs := []any{slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.Int("status", rw.statusCode), slog.String("remote_ip", r.RemoteAddr)}
	if rw.statusCode >= 400 {
		attrs = append(attrs, slog.String("body", string(rw.errorBody)))
		providers.Log().Error("request", attrs...)
	} else {
		providers.Log().Info("request", attrs...)
	}
}

func GetProviders(r *http.Request) config.Providers {
	providers, ok := r.Context().Value(providersCtxKey{}).(config.Providers)
	if !ok {
		return nil
	}
	return providers
}
