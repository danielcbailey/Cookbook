package apicommon

import (
	"encoding/json"
	"net/http"

	"github.com/danielcbailey/Cookbook/core/config"
)

const (
	MiB = 1024 * 1024
	KiB = 1024
)

func ExpectedMethod(w http.ResponseWriter, r *http.Request, endpoint string, method string) bool {
	if r.Method == method {
		return true
	}

	http.Error(w, endpoint+" endpoint requires method "+method, http.StatusMethodNotAllowed)
	return false
}

func MustHaveProvidersAndUser(r *http.Request) config.Providers {
	providers := GetProviders(r)
	if providers == nil || providers.User() == nil {
		panic("endpoint must have providers and user")
	}

	return providers
}

func DecodeRequest[T any](w http.ResponseWriter, r *http.Request, maxSize int64) (T, bool) {
	limited := http.MaxBytesReader(w, r.Body, maxSize)

	var ret T
	err := json.NewDecoder(limited).Decode(&ret)
	if err != nil {
		http.Error(w, "failed to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return ret, false
	}

	return ret, true
}

func WriteJSON(w http.ResponseWriter, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(payload)
}
