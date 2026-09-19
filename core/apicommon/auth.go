package apicommon

import (
	"context"
	"crypto/rand"
	"math/big"
	"net/http"
	"time"

	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/danielcbailey/Cookbook/core/models"
)

type authHandler struct {
	next http.Handler
}

func AuthRequired(next http.Handler) http.Handler {
	return &authHandler{next: next}
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := getToken(r)
	providers := GetProviders(r)
	if providers == nil {
		panic("expected providers set")
	}

	// Checking cache
	var user models.User
	err := providers.Cache().GetInterface(r.Context(), sessionKey(token), &user)
	if err != nil {
		if providers.Cache().ErrIsNotFound(err) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		providers.Log().Error("failed to authorize request", "error", err)

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	providers = providers.WithUser(&user)
	ctx := context.WithValue(r.Context(), providersCtxKey{}, providers)
	h.next.ServeHTTP(w, r.WithContext(ctx))
}

func getToken(r *http.Request) string {
	if cookie, err := r.Cookie("session"); err == nil {
		return cookie.Value
	}
	if h := r.Header.Get("Authorization"); len(h) > 7 && h[:7] == "Bearer " {
		return h[7:]
	}
	return ""
}

func CreateUserToken(ctx context.Context, providers config.Providers, user *models.User, long bool) (string, time.Time, error) {
	token, err := secureRandomString(32)
	if err != nil {
		return "", time.Time{}, err
	}

	expirySeconds := providers.Config().TokenExpirySeconds
	if long {
		expirySeconds = providers.Config().TokenLongExpirySeconds
	}

	if expirySeconds == 0 && !long {
		expirySeconds = 8 * 60 * 60 // 8 hours
	} else if expirySeconds == 0 {
		expirySeconds = 30 * 24 * 60 * 60 // 30 days
	}

	expiryDuration := time.Duration(expirySeconds) * time.Second

	return token, time.Now().Add(expiryDuration), providers.Cache().SetInterface(ctx, sessionKey(token), user, expiryDuration)
}

func secureRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range b {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}

func sessionKey(token string) string {
	return "sessions:" + token
}
