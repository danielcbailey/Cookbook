package recipeapi

import (
	"net/http"

	"github.com/danielcbailey/Cookbook/core/apicommon"
)

func Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/recipe/importweb", apicommon.AuthRequired(http.HandlerFunc(handleRecipeWebImport)))
	mux.Handle("/api/v1/recipe/importphotos", apicommon.AuthRequired(http.HandlerFunc(handleRecipePhotoImport)))
	return mux
}
