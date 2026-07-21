package apicommon

import (
	"net/http"

	"github.com/danielcbailey/Cookbook/internal/models"
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
	err := providers.Cache().GetInterface(r.Context(), "sessions:"+token, &user)
	if err != nil {
		if providers.Cache().ErrIsNotFound(err) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		providers.Log().Error("failed to authorize request", "error", err)

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
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
