package userapi

import (
	"net/http"

	"github.com/danielcbailey/Cookbook/core/apicommon"
)

func Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/user/login", handleUserLogin)
	mux.Handle("/api/v1/user/me", apicommon.AuthRequired(http.HandlerFunc(handleGetUserInfo)))
	return mux
}
