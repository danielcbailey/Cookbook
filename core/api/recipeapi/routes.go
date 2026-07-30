package recipeapi

import (
	"net/http"

	"github.com/danielcbailey/Cookbook/core/apicommon"
)

func Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/recipe/importweb", apicommon.AuthRequired(http.HandlerFunc(handleRecipeWebImport)))
	mux.Handle("/api/v1/recipe/importphotos", apicommon.AuthRequired(http.HandlerFunc(handleRecipePhotoImport)))
	mux.Handle("/api/v1/recipe/save", apicommon.AuthRequired(http.HandlerFunc(handleRecipeSave)))
	mux.Handle("/api/v1/recipe/delete", apicommon.AuthRequired(http.HandlerFunc(handleRecipeDelete)))

	mux.Handle("/api/v1/recipe/list", apicommon.AuthRequired(http.HandlerFunc(handleListRecipes)))
	mux.Handle("/api/v1/recipe/listbycategory", apicommon.AuthRequired(http.HandlerFunc(handleListRecipesByCategory)))
	mux.Handle("/api/v1/recipe/listbyprotein", apicommon.AuthRequired(http.HandlerFunc(handleListRecipesByProtein)))
	mux.Handle("/api/v1/recipe/listbymeal", apicommon.AuthRequired(http.HandlerFunc(handleListRecipesByMeal)))
	mux.Handle("/api/v1/recipe/search", apicommon.AuthRequired(http.HandlerFunc(handleListRecipesSearch)))

	mux.Handle("/api/v1/recipe/categories", apicommon.AuthRequired(http.HandlerFunc(handleListCategories)))
	mux.Handle("/api/v1/recipe/proteins", apicommon.AuthRequired(http.HandlerFunc(handleListProteins)))
	mux.Handle("/api/v1/recipe/mealtimes", apicommon.AuthRequired(http.HandlerFunc(handleListMealtimes)))
	return mux
}
