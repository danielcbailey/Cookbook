package recipeapi

import (
	"log/slog"
	"net/http"

	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/internal/recipes"
)

func handleRecipeDelete(w http.ResponseWriter, r *http.Request) {
	if !apicommon.ExpectedMethod(w, r, "recipe web import", http.MethodPost) {
		return
	}

	reqObj, ok := apicommon.DecodeRequest[RecipeID](w, r, 1*apicommon.KiB)
	if !ok {
		return
	}

	p := apicommon.MustHaveProvidersAndUser(r)

	err := recipes.DeleteRecipe(r.Context(), p, reqObj.ID)
	if err != nil {
		placeholder := "internal server error"
		sanitized := apicommon.SanitizeError(err, placeholder)
		if placeholder == sanitized {
			p.Log().Error("failed to save recipe", slog.Any("error", err))
			http.Error(w, placeholder, http.StatusInternalServerError)
			return
		}

		http.Error(w, sanitized, http.StatusBadRequest)
		return
	}
}
