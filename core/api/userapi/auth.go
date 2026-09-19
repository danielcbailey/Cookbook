package userapi

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/pkg/database"
)

type UserLoginRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	LongLived bool   `json:"long_lived"`
}

type UserLoginResponse struct {
	Token  string `json:"token"`
	Expiry int64  `json:"expiry"`
}

func handleUserLogin(w http.ResponseWriter, r *http.Request) {
	if !apicommon.ExpectedMethod(w, r, "login", http.MethodPost) {
		return
	}

	parsedReq, ok := apicommon.DecodeRequest[UserLoginRequest](w, r, 4*apicommon.KiB)
	if !ok {
		return
	}

	p := apicommon.GetProviders(r)
	if p == nil {
		panic("expected providers set for login endpoint")
	}

	tx, err := p.DB().NewTransaction(r.Context())
	if err != nil {
		p.Log().Error("failed to create new DB transaction", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Read only, so nothing to commit

	user, err := tx.GetUserByPasswordHash(parsedReq.Email, getPasswordHash(parsedReq.Password))
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		p.Log().Error("failed to obtain user record for login", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	token, expiry, err := apicommon.CreateUserToken(r.Context(), p, user, parsedReq.LongLived)
	if err != nil {
		p.Log().Error("failed to create token for user login", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	responseObj := UserLoginResponse{
		Token:  token,
		Expiry: expiry.Unix(),
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		MaxAge:   int(time.Until(expiry).Seconds()),
		HttpOnly: true,
	})

	err = apicommon.WriteJSON(w, &responseObj)
	if err != nil {
		p.Log().Warn("failed to send login response", slog.Any("error", err))
	}
}

func getPasswordHash(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}
