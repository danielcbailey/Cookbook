package recipeapi

import (
	"log/slog"
	"net/http"

	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/internal/ai"
	"github.com/danielcbailey/Cookbook/internal/recipes"
	"github.com/danielcbailey/Cookbook/pkg/database"
)

func handleListRecipeCommon(w http.ResponseWriter, r *http.Request, listFn func(tx database.Transaction, userID int64) ([]*models.RecipeListing, error)) {
	if !apicommon.ExpectedMethod(w, r, "list recipes", http.MethodGet) {
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

	listings, err := listFn(tx, p.User().ID)
	if err != nil {
		p.Log().Error("failed to query database", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = recipes.RecipeListingPrepareImageURLs(r.Context(), p, listings)
	if err != nil {
		p.Log().Error("failed to prepare image URLs", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = apicommon.WriteJSON(w, listings)
	if err != nil {
		p.Log().Warn("failed to write response", slog.Any("error", err))
	}
}

func handleListRecipes(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	offset, offOk := apicommon.ParsePositiveIntParam(w, params, "offset")
	limit, limOk := apicommon.ParsePositiveIntParam(w, params, "limit")
	if !offOk || !limOk {
		return
	}

	handleListRecipeCommon(w, r, func(tx database.Transaction, userID int64) ([]*models.RecipeListing, error) {
		return tx.ListRecipesByUserID(userID, offset, limit)
	})
}

func handleListRecipesByCategory(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	category := params.Get("category")
	limit, limOk := apicommon.ParsePositiveIntParam(w, params, "limit")
	if !limOk {
		return
	}

	handleListRecipeCommon(w, r, func(tx database.Transaction, userID int64) ([]*models.RecipeListing, error) {
		return tx.ListRecipesByCategory(userID, category, limit)
	})
}

func handleListRecipesByProtein(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	protein := params.Get("protein")
	limit, limOk := apicommon.ParsePositiveIntParam(w, params, "limit")
	if !limOk {
		return
	}

	handleListRecipeCommon(w, r, func(tx database.Transaction, userID int64) ([]*models.RecipeListing, error) {
		return tx.ListRecipesByProtein(userID, protein, limit)
	})
}

func handleListRecipesByMeal(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	meal := params.Get("mealtime")
	limit, limOk := apicommon.ParsePositiveIntParam(w, params, "limit")
	if !limOk {
		return
	}

	handleListRecipeCommon(w, r, func(tx database.Transaction, userID int64) ([]*models.RecipeListing, error) {
		return tx.ListRecipesByMeal(userID, meal, limit)
	})
}

func handleListRecipesByTag(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	tag := params.Get("tag")
	limit, limOk := apicommon.ParsePositiveIntParam(w, params, "limit")
	if !limOk {
		return
	}

	handleListRecipeCommon(w, r, func(tx database.Transaction, userID int64) ([]*models.RecipeListing, error) {
		return tx.ListRecipesByTag(userID, tag, limit)
	})
}

func handleListRecipesSearch(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	query := params.Get("query")
	limit, limOk := apicommon.ParsePositiveIntParam(w, params, "limit")
	if !limOk {
		return
	}

	p := apicommon.MustHaveProvidersAndUser(r)
	embeddingsEnabled := p.Config().SemanticSearchEnabled

	handleListRecipeCommon(w, r, func(tx database.Transaction, userID int64) ([]*models.RecipeListing, error) {
		if embeddingsEnabled {
			embedding, err := ai.ConvertToEmbedding(r.Context(), p, query)
			if err != nil {
				return nil, err
			}

			return tx.ListRecipesBySemanticSimilarity(userID, embedding, limit)
		} else {
			return tx.ListRecipesByTitleSearch(userID, query, limit)
		}
	})
}

func handleListOptionsCommon(w http.ResponseWriter, r *http.Request, listFn func(tx database.Transaction, userID int64) ([]string, error)) {
	if !apicommon.ExpectedMethod(w, r, "list recipe options", http.MethodGet) {
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

	listings, err := listFn(tx, p.User().ID)
	if err != nil {
		p.Log().Error("failed to query database", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = apicommon.WriteJSON(w, listings)
	if err != nil {
		p.Log().Warn("failed to write response", slog.Any("error", err))
	}
}

func handleListCategories(w http.ResponseWriter, r *http.Request) {
	handleListOptionsCommon(w, r, func(tx database.Transaction, userID int64) ([]string, error) {
		return tx.ListRecipeCategories(userID)
	})
}

func handleListProteins(w http.ResponseWriter, r *http.Request) {
	handleListOptionsCommon(w, r, func(tx database.Transaction, userID int64) ([]string, error) {
		return tx.ListRecipeProteins(userID)
	})
}

func handleListMealtimes(w http.ResponseWriter, r *http.Request) {
	handleListOptionsCommon(w, r, func(tx database.Transaction, userID int64) ([]string, error) {
		return tx.ListRecipeMealtimes(userID)
	})
}

func handleListTags(w http.ResponseWriter, r *http.Request) {
	handleListOptionsCommon(w, r, func(tx database.Transaction, userID int64) ([]string, error) {
		return tx.ListRecipeTags(userID)
	})
}
