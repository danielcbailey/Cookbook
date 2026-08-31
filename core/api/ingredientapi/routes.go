package ingredientapi

import (
	"net/http"

	"github.com/danielcbailey/Cookbook/core/apicommon"
)

func Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/ingredient/list", apicommon.AuthRequired(http.HandlerFunc(handleListIngredients)))
	return mux
}
