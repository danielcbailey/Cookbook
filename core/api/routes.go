package api

import (
	"net/http"

	"github.com/danielcbailey/Cookbook/core/api/userapi"
)

func Handlers() map[string]http.Handler {
	return map[string]http.Handler{
		"/api/v1/user": userapi.Routes(),
	}
}
