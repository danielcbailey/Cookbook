package recipeapi

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/internal/recipes"
	"github.com/danielcbailey/Cookbook/pkg/database"
)

func handleGetRecipe(w http.ResponseWriter, r *http.Request) {
	if !apicommon.ExpectedMethod(w, r, "get user info", http.MethodGet) {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "failed to parse id: "+err.Error(), http.StatusBadRequest)
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

	recipe, err := tx.GetRecipeByID(id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, "invalid recipe ID", http.StatusNotFound)
			return
		}
	}

	err = recipes.RecipePrepareImageURLs(r.Context(), p, recipe)
	if err != nil {
		p.Log().Error("failed to prepare image URLs", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = apicommon.WriteJSON(w, recipe)
	if err != nil {
		p.Log().Warn("failed to write response", slog.Any("error", err))
	}
}
