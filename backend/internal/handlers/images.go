package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
)

var imageSlots = make(chan struct{}, 2)
var activityImageName = regexp.MustCompile(`^[a-f0-9]{32}\.avif$`)

// Only raster pixels with the exact output dimensions reach the encoder.
func (h *Handler) AdminUploadActivityImage(w http.ResponseWriter, r *http.Request) {
	select {
	case imageSlots <- struct{}{}:
		defer func() { <-imageSlots }()
	default:
		writeError(w, http.StatusTooManyRequests, "Hay imágenes procesándose. Inténtalo en unos segundos")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "La imagen recortada supera el límite de 12 MB")
		return
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width != 1080 || cfg.Height != 1350 {
		writeError(w, http.StatusUnprocessableEntity, "Recorta la imagen a 1080 × 1350 antes de subirla")
		return
	}
	pixels, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "No pudimos leer la imagen")
		return
	}
	dir := filepath.Join(h.Cfg.UploadDir, "activities")
	if err := os.MkdirAll(dir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo preparar el almacenamiento")
		return
	}
	work, err := os.MkdirTemp(dir, ".processing-")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo preparar la imagen")
		return
	}
	defer os.RemoveAll(work)
	input := filepath.Join(work, "input.png")
	file, err := os.Create(input)
	if err == nil {
		err = png.Encode(file, pixels)
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo procesar la imagen")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()
	output := filepath.Join(work, "output.avif")
	if err := exec.CommandContext(ctx, "avifenc", "--jobs", "2", "--speed", "8", "--min", "20", "--max", "30", input, output).Run(); err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo convertir la imagen a AVIF. Inténtalo de nuevo")
		return
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo guardar la imagen")
		return
	}
	name := hex.EncodeToString(random[:]) + ".avif"
	if err := os.Rename(output, filepath.Join(dir, name)); err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo guardar la imagen")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"imagen_url": "/media/activities/" + name, "width": 1080, "height": 1350})
}

func (h *Handler) ActivityImage(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "file")
	if !activityImageName.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	path := filepath.Join(h.Cfg.UploadDir, "activities", name)
	if _, err := os.Stat(path); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/avif")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, path)
}
