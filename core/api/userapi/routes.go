package userapi

import "net/http"

func Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/user/login", handleUserLogin)
	return mux
}
