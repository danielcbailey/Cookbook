package recipes

import (
	"bytes"
	"context"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/internal/ai"
)

const maxScrapeSize = 1 * apicommon.MiB
const maxPhotoArea = 10_000_000 // pixels
const maxPhotos = 3

func ScrapeRecipeWeb(ctx context.Context, providers config.Providers, url string) (*models.Recipe, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, apicommon.NewUserFacingError("failed to get site contents: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return nil, apicommon.NewUserFacingError("website responded with invalid status, expected OK. Got %s (%d)", resp.Status, resp.StatusCode)
	} else if cType := resp.Header.Get("Content-Type"); strings.SplitN(cType, ";", 2)[0] != "text/html" {
		return nil, apicommon.NewUserFacingError("invalid website content type, expected text/html. Got %s", cType)
	}

	limited := io.LimitReader(resp.Body, maxScrapeSize+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}

	if len(body) > maxScrapeSize {
		return nil, apicommon.NewUserFacingError("website content is too long. Max size is %d KiB", maxScrapeSize/apicommon.KiB)
	}

	return ai.ExtractRecipeFromHTML(ctx, providers, string(body))
}

func ScrapeRecipePhotos(ctx context.Context, providers config.Providers, files []ai.FileAttachment) (*models.Recipe, error) {
	if len(files) > maxPhotos {
		return nil, apicommon.NewUserFacingError("too many photos, maximum of %d allowed", maxPhotos)
	}

	for i, file := range files {
		if file.MimeType != "image/jpeg" && file.MimeType != "image/png" {
			return nil, apicommon.NewUserFacingError("photo %d has unsupported type %s, only JPEG and PNG are allowed", i+1, file.MimeType)
		}

		cfg, _, err := image.DecodeConfig(bytes.NewReader(file.Content))
		if err != nil {
			return nil, apicommon.NewUserFacingError("photo %d is not a valid image: %v", i+1, err)
		}

		if cfg.Width*cfg.Height > maxPhotoArea {
			return nil, apicommon.NewUserFacingError("photo %d exceeds the maximum image size of %d megapixels", i+1, maxPhotoArea/1_000_000)
		}
	}

	return ai.ExtractRecipeFromPhotos(ctx, providers, files)
}
