package ingredientapi

import (
	"log/slog"
	"net/http"

	"github.com/danielcbailey/Cookbook/core/apicommon"
)

func handleListIngredients(w http.ResponseWriter, r *http.Request) {
	if !apicommon.ExpectedMethod(w, r, "list ingredients", http.MethodGet) {
		return
	}

	params := r.URL.Query()
	offset, offOk := apicommon.ParsePositiveIntParam(w, params, "offset")
	limit, limOk := apicommon.ParsePositiveIntParam(w, params, "limit")
	if !offOk || !limOk {
		return
	}

	p := apicommon.MustHaveProvidersAndUser(r)
	tx, err := p.DB().NewTransaction(r.Context())
	if err != nil {
		p.Log().Error("failed to start database transaction", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Read only, so nothing to commit

	ingredients, err := tx.ListIngredients(p.User().ID, offset, limit)
	if err != nil {
		p.Log().Error("failed to query database", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// The local backend carries embeddings on its models while postgres never
	// selects them. Drop them so both backends return the same payload and the
	// response stays small.
	for _, ing := range ingredients {
		ing.Embedding = nil
	}

	err = apicommon.WriteJSON(w, ingredients)
	if err != nil {
		p.Log().Warn("failed to write response", slog.Any("error", err))
	}
}
