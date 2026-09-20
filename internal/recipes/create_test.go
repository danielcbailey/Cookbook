package recipes

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"net/url"
	"testing"
)

// pngDataURL builds a data URL whose base64 payload is long and noisy enough to
// contain '+' and '/' characters, which is what a real uploaded photo looks like.
func pngDataURL(t *testing.T) string {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 7), G: uint8(y * 5), B: uint8(x ^ y), A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode fixture image: %v", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	if !bytes.ContainsAny([]byte(encoded), "+/") {
		t.Fatalf("fixture payload lacks the '+' and '/' characters the test is about")
	}

	return "data:image/png;base64," + encoded
}

func TestDecodeEmbeddedImage(t *testing.T) {
	dataURL := pngDataURL(t)

	cases := []struct {
		name  string
		input string
	}{
		{"plain", dataURL},
		{"percent encoded", url.PathEscape(dataURL)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, isEmbedded, err := decodeEmbeddedImage(tc.input)
			if err != nil {
				t.Fatalf("decodeEmbeddedImage returned an error: %v", err)
			}
			if !isEmbedded {
				t.Fatal("expected the data URL to be reported as embedded")
			}
			if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
				t.Fatalf("decoded payload is not a valid image: %v", err)
			}
		})
	}
}

func TestDecodeEmbeddedImageIgnoresRemoteURLs(t *testing.T) {
	_, isEmbedded, err := decodeEmbeddedImage("https://example.com/photo.jpg")
	if err != nil {
		t.Fatalf("decodeEmbeddedImage returned an error: %v", err)
	}
	if isEmbedded {
		t.Fatal("a remote URL must not be treated as embedded")
	}
}
