package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
	"unicode/utf8"
)

type GalleryImage struct {
	ID          string `json:"id"`
	ImagenURL   string `json:"imagen_url"`
	Descripcion string `json:"descripcion"`
	Posicion    int    `json:"posicion"`
	Activa      bool   `json:"activa"`
}

func (h *Handler) ListGallery(w http.ResponseWriter, r *http.Request)      { h.listGallery(w, r, false) }
func (h *Handler) AdminListGallery(w http.ResponseWriter, r *http.Request) { h.listGallery(w, r, true) }
func (h *Handler) listGallery(w http.ResponseWriter, r *http.Request, admin bool) {
	query := `SELECT id::text, imagen_url, descripcion, posicion, activa FROM gallery_images`
	if !admin {
		query += ` WHERE activa=TRUE`
	}
	query += ` ORDER BY posicion ASC, created_at ASC`
	rows, err := h.DB.QueryContext(r.Context(), query)
	if err != nil {
		writeError(w, 500, "No se pudo cargar la galería")
		return
	}
	defer rows.Close()
	images := []GalleryImage{}
	for rows.Next() {
		var item GalleryImage
		if err := rows.Scan(&item.ID, &item.ImagenURL, &item.Descripcion, &item.Posicion, &item.Activa); err != nil {
			writeError(w, 500, "No se pudo leer la galería")
			return
		}
		images = append(images, item)
	}
	if rows.Err() != nil {
		writeError(w, 500, "No se pudo leer la galería")
		return
	}
	writeJSON(w, 200, images)
}
func (h *Handler) AdminCreateGallery(w http.ResponseWriter, r *http.Request) {
	h.saveGallery(w, r, true)
}
func (h *Handler) AdminUpdateGallery(w http.ResponseWriter, r *http.Request) {
	h.saveGallery(w, r, false)
}
func (h *Handler) saveGallery(w http.ResponseWriter, r *http.Request, create bool) {
	var item GalleryImage
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	if json.NewDecoder(r.Body).Decode(&item) != nil {
		writeError(w, 400, "Datos inválidos")
		return
	}
	item.Descripcion = strings.TrimSpace(item.Descripcion)
	name := strings.TrimPrefix(item.ImagenURL, "/media/activities/")
	if !strings.HasPrefix(item.ImagenURL, "/media/activities/") || !activityImageName.MatchString(name) || utf8.RuneCountInString(item.Descripcion) > 180 || item.Posicion < 0 || item.Posicion > 10000 {
		writeError(w, 422, "Sube una imagen válida, usa hasta 180 caracteres y un orden entre 0 y 10000")
		return
	}
	if create {
		err := h.DB.QueryRowContext(r.Context(), `INSERT INTO gallery_images (imagen_url,descripcion,posicion,activa) VALUES ($1,$2,$3,$4) RETURNING id::text`, item.ImagenURL, item.Descripcion, item.Posicion, item.Activa).Scan(&item.ID)
		if err != nil {
			writeError(w, 500, "No se pudo guardar la foto")
			return
		}
		writeJSON(w, 201, item)
		return
	}
	result, err := h.DB.ExecContext(r.Context(), `UPDATE gallery_images SET imagen_url=$1,descripcion=$2,posicion=$3,activa=$4 WHERE id=$5`, item.ImagenURL, item.Descripcion, item.Posicion, item.Activa, chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, 500, "No se pudo actualizar la foto")
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		writeError(w, 404, "Foto no encontrada")
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// Remove the gallery entry; uploaded media may also be referenced by another entry.
func (h *Handler) AdminDeleteGallery(w http.ResponseWriter, r *http.Request) {
	result, err := h.DB.ExecContext(r.Context(), `DELETE FROM gallery_images WHERE id=$1`, chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo eliminar la foto")
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "Foto no encontrada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
