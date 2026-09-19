package api

import "net/http"

func handleLiveCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func handleReadyCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
