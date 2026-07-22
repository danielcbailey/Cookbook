package recipeapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/internal/ai"
	"github.com/danielcbailey/Cookbook/internal/recipes"
)

type RecipeWebImportRequest struct {
	URL string
}

func handleRecipeWebImport(w http.ResponseWriter, r *http.Request) {
	if !apicommon.ExpectedMethod(w, r, "recipe web import", http.MethodPost) {
		return
	}

	reqObj, ok := apicommon.DecodeRequest[RecipeWebImportRequest](w, r, 8*apicommon.KiB)
	if !ok {
		return
	}

	p := apicommon.MustHaveProvidersAndUser(r)

	if !p.Config().RecipeImportEnabled {
		http.Error(w, "recipe import not enabled", http.StatusNotImplemented)
		return
	} else if !checkUserImportLimit(p.User()) {
		http.Error(w, "exceeded recipe import limit", http.StatusForbidden)
		return
	}

	recipe, err := recipes.ScrapeRecipeWeb(r.Context(), p, reqObj.URL)
	if err != nil {
		placeholder := "internal server error"
		sanitized := apicommon.SanitizeError(err, placeholder)
		if placeholder == sanitized {
			p.Log().Error("failed to scrape recipe: %v", slog.Any("error", err))
			http.Error(w, placeholder, http.StatusInternalServerError)
			return
		}

		http.Error(w, sanitized, http.StatusBadRequest)
		return
	}

	// Updating user usage of recipe extraction
	tx, err := p.DB().NewTransaction(r.Context())
	if err != nil {
		p.Log().Error("failed to create DB transaction", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	err = tx.UpdateUserUsage(p.User().ID, 0, 1)
	if err != nil {
		p.Log().Error("failed to update user recipe import usage", slog.Any("error", err))
		tx.Rollback()
	} else {
		tx.Commit()
	}

	err = json.NewEncoder(w).Encode(recipe)
	if err != nil {
		p.Log().Warn("failed to write response", slog.Any("error", err))
	}
}

const maxPhotoFormSize = 20 * apicommon.MiB

func handleRecipePhotoImport(w http.ResponseWriter, r *http.Request) {
	if !apicommon.ExpectedMethod(w, r, "recipe photo import", http.MethodPost) {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPhotoFormSize)
	if err := r.ParseMultipartForm(maxPhotoFormSize); err != nil {
		http.Error(w, "form too large or invalid, max size is 20 MiB", http.StatusBadRequest)
		return
	}

	p := apicommon.MustHaveProvidersAndUser(r)

	if !p.Config().RecipeImportEnabled {
		http.Error(w, "recipe import not enabled", http.StatusNotImplemented)
		return
	} else if !checkUserImportLimit(p.User()) {
		http.Error(w, "exceeded recipe import limit", http.StatusForbidden)
		return
	}

	multipartFiles := r.MultipartForm.File["photos"]
	files := make([]ai.FileAttachment, 0, len(multipartFiles))
	for _, fh := range multipartFiles {
		f, err := fh.Open()
		if err != nil {
			http.Error(w, "failed to read uploaded file", http.StatusBadRequest)
			return
		}

		content, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			http.Error(w, "failed to read uploaded file", http.StatusBadRequest)
			return
		}

		files = append(files, ai.FileAttachment{
			MimeType: fh.Header.Get("Content-Type"),
			Content:  content,
		})
	}

	recipe, err := recipes.ScrapeRecipePhotos(r.Context(), p, files)
	if err != nil {
		placeholder := "internal server error"
		sanitized := apicommon.SanitizeError(err, placeholder)
		if placeholder == sanitized {
			p.Log().Error("failed to extract recipe from photos", slog.Any("error", err))
			http.Error(w, placeholder, http.StatusInternalServerError)
			return
		}

		http.Error(w, sanitized, http.StatusBadRequest)
		return
	}

	tx, err := p.DB().NewTransaction(r.Context())
	if err != nil {
		p.Log().Error("failed to create DB transaction", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	err = tx.UpdateUserUsage(p.User().ID, 0, 1)
	if err != nil {
		p.Log().Error("failed to update user recipe import usage", slog.Any("error", err))
		tx.Rollback()
	} else {
		tx.Commit()
	}

	err = json.NewEncoder(w).Encode(recipe)
	if err != nil {
		p.Log().Warn("failed to write response", slog.Any("error", err))
	}
}

func checkUserImportLimit(user *models.User) bool {
	usage := user.CurrentMonthlyRecipeExtraction
	now := time.Now()
	if now.Year() != user.LastUsageReset.Year() || now.Month() != user.LastUsageReset.Month() {
		usage = 0
	}
	return usage < user.MaxMonthlyRecipeExtraction
}
