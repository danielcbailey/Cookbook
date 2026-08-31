package userapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/danielcbailey/Cookbook/core/apicommon"
)

func handleGetUserInfo(w http.ResponseWriter, r *http.Request) {
	if !apicommon.ExpectedMethod(w, r, "get user info", http.MethodGet) {
		return
	}

	p := apicommon.MustHaveProvidersAndUser(r)
	user := *p.User()
	user.LastUsageReset = time.Time{}
	user.PasswordHash = ""

	err := apicommon.WriteJSON(w, &user)
	if err != nil {
		p.Log().Warn("failed to write response", slog.Any("error", err))
	}
}
