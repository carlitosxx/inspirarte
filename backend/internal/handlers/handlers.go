package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"inspirarte-backend/internal/auth"
	"inspirarte-backend/internal/config"
)

var lima, _ = time.LoadLocation("America/Lima")

// Reglas del negocio (datos del cliente)
const (
	duracionCafeteria   = 150 // 2 h 30
	duracionCelebracion = 180 // 3 h
	toleranciaMin       = 15
)

type Handler struct {
	DB            *sql.DB
	Cfg           config.Config
	AdminPassHash []byte
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---------------------------------------------------------------------------
// Salud
// ---------------------------------------------------------------------------

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "inspirarte-api"})
}

// ---------------------------------------------------------------------------
// Actividades (público)
// ---------------------------------------------------------------------------

type Activity struct {
	ID               string  `json:"id"`
	Titulo           string  `json:"titulo"`
	Descripcion      string  `json:"descripcion,omitempty"`
	ImagenURL        string  `json:"imagen_url,omitempty"`
	FechaHora        string  `json:"fecha_hora"`
	Precio           float64 `json:"precio"`
	CuposTotales     int     `json:"cupos_totales"`
	CuposDisponibles int     `json:"cupos_disponibles"`
	Agotado          bool    `json:"agotado"`
	Activa           bool    `json:"activa,omitempty"`
}

func scanActivity(row interface{ Scan(...any) error }) (Activity, error) {
	var a Activity
	var fechaHora time.Time
	var descripcion, imagen sql.NullString
	err := row.Scan(&a.ID, &a.Titulo, &descripcion, &imagen, &fechaHora, &a.Precio, &a.CuposTotales, &a.CuposDisponibles)
	if err != nil {
		return a, err
	}
	a.Descripcion = descripcion.String
	a.ImagenURL = imagen.String
	a.FechaHora = fechaHora.Format(time.RFC3339)
	a.Agotado = a.CuposDisponibles <= 0
	return a, nil
}

const activityCols = `id::text, titulo, descripcion, imagen_url, fecha_hora, precio, cupos_totales, cupos_disponibles`

func (h *Handler) ListActivities(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT ` + activityCols + ` FROM activities WHERE activa = TRUE AND fecha_hora > now() ORDER BY fecha_hora ASC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudieron cargar las actividades")
		return
	}
	defer rows.Close()
	acts := []Activity{}
	for rows.Next() {
		a, err := scanActivity(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error leyendo actividades")
			return
		}
		acts = append(acts, a)
	}
	writeJSON(w, http.StatusOK, acts)
}

type BookRequest struct {
	Nombre   string `json:"nombre"`
	Whatsapp string `json:"whatsapp"`
	Cupos    int    `json:"cupos"`
}

// BookActivity aparta cupos con descuento atómico: imposible hacer overbooking
// aunque dos personas reserven al mismo tiempo.
func (h *Handler) BookActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}
	req.Nombre = strings.TrimSpace(req.Nombre)
	req.Whatsapp = strings.TrimSpace(req.Whatsapp)
	if req.Nombre == "" || req.Whatsapp == "" {
		writeError(w, http.StatusUnprocessableEntity, "Nombre y WhatsApp son obligatorios")
		return
	}
	if req.Cupos < 1 {
		req.Cupos = 1
	}

	tx, err := h.DB.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo procesar la reserva")
		return
	}
	defer tx.Rollback()

	var restantes int
	err = tx.QueryRow(
		`UPDATE activities SET cupos_disponibles = cupos_disponibles - $1
		 WHERE id = $2 AND activa = TRUE AND fecha_hora > now() AND cupos_disponibles >= $1
		 RETURNING cupos_disponibles`,
		req.Cupos, id,
	).Scan(&restantes)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusConflict, "Cupos agotados ✨")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo apartar el cupo")
		return
	}
	if _, err := tx.Exec(
		`INSERT INTO activity_bookings (activity_id, nombre, whatsapp, cupos) VALUES ($1, $2, $3, $4)`,
		id, req.Nombre, req.Whatsapp, req.Cupos,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo registrar tu reserva")
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo confirmar tu reserva")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"ok": true, "cupos_restantes": restantes,
		"mensaje": "¡Cupo apartado! 💜 Te escribiremos por WhatsApp para confirmar.",
	})
}

// ---------------------------------------------------------------------------
// Reservas de mesa (público)
// ---------------------------------------------------------------------------

type ReservationRequest struct {
	Nombre        string `json:"nombre"`
	Whatsapp      string `json:"whatsapp"`
	Fecha         string `json:"fecha"` // YYYY-MM-DD
	Hora          string `json:"hora"`  // HH:MM
	Personas      int    `json:"personas"`
	Ubicacion     string `json:"ubicacion"` // salon | mampara | barra
	Motivo        string `json:"motivo"`    // cafeteria | cumpleanos | aniversario | reunion | experiencia | otro
	MotivoDetalle string `json:"motivo_detalle"`
}

var ubicacionesValidas = map[string]bool{"salon": true, "mampara": true, "barra": true}
var motivosValidos = map[string]bool{
	"cafeteria": true, "cumpleanos": true, "aniversario": true,
	"reunion": true, "experiencia": true, "otro": true,
}

// dentroDeHorario valida la hora contra el horario de atención del cliente.
func dentroDeHorario(fecha time.Time, hora time.Time) bool {
	mins := hora.Hour()*60 + hora.Minute()
	switch fecha.In(lima).Weekday() {
	case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday:
		return mins >= 17*60 && mins <= 22*60 // 5:00 p. m. – 10:00 p. m.
	case time.Saturday:
		return mins >= 10*60 && mins <= 22*60 // 10:00 a. m. – 10:00 p. m.
	default:
		return false // domingo: se valida aparte
	}
}

func esDomingo(fecha time.Time) bool { return fecha.In(lima).Weekday() == time.Sunday }

func esCelebracionOEspecial(motivo string, personas int) bool {
	switch motivo {
	case "cumpleanos", "aniversario", "reunion", "experiencia":
		return true
	}
	return personas >= 6 // reserva grupal
}

func duracionPorMotivo(motivo string) int {
	switch motivo {
	case "cumpleanos", "aniversario", "reunion":
		return duracionCelebracion
	default:
		return duracionCafeteria
	}
}

func (h *Handler) CreateReservation(w http.ResponseWriter, r *http.Request) {
	var req ReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}
	req.Nombre = strings.TrimSpace(req.Nombre)
	req.Whatsapp = strings.TrimSpace(req.Whatsapp)
	req.MotivoDetalle = strings.TrimSpace(req.MotivoDetalle)

	if req.Nombre == "" || req.Whatsapp == "" {
		writeError(w, http.StatusUnprocessableEntity, "Nombre y WhatsApp son obligatorios")
		return
	}
	if req.Personas < 1 {
		writeError(w, http.StatusUnprocessableEntity, "El número de personas debe ser al menos 1")
		return
	}
	if !ubicacionesValidas[req.Ubicacion] {
		writeError(w, http.StatusUnprocessableEntity, "Ubicación inválida")
		return
	}
	if req.Ubicacion == "mampara" && req.Personas > 2 {
		writeError(w, http.StatusUnprocessableEntity, "La mesa junto a la mampara es para un máximo de 2 personas")
		return
	}
	if !motivosValidos[req.Motivo] {
		writeError(w, http.StatusUnprocessableEntity, "Motivo de visita inválido")
		return
	}
	if req.Motivo == "otro" && req.MotivoDetalle == "" {
		writeError(w, http.StatusUnprocessableEntity, "Cuéntanos el motivo de tu visita para recibirte mejor 💜")
		return
	}

	fecha, err := time.ParseInLocation("2006-01-02", req.Fecha, lima)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Fecha inválida")
		return
	}
	hoy := time.Now().In(lima).Truncate(24 * time.Hour)
	if fecha.Before(hoy) {
		writeError(w, http.StatusUnprocessableEntity, "La fecha no puede ser anterior a hoy")
		return
	}
	hora, err := time.Parse("15:04", req.Hora)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Hora inválida")
		return
	}
	if esDomingo(fecha) {
		if !esCelebracionOEspecial(req.Motivo, req.Personas) {
			writeError(w, http.StatusUnprocessableEntity,
				"Los domingos atendemos únicamente con reserva previa para grupos o celebraciones especiales ✨")
			return
		}
	} else if !dentroDeHorario(fecha, hora) {
		writeError(w, http.StatusUnprocessableEntity,
			"Esa hora está fuera de nuestro horario de atención (lun–vie 5:00 p. m. – 10:00 p. m., sáb 10:00 a. m. – 10:00 p. m.)")
		return
	}

	var id string
	err = h.DB.QueryRow(
		`INSERT INTO reservations (nombre, whatsapp, fecha, hora, personas, ubicacion, motivo, motivo_detalle, duracion_min)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8,''), $9) RETURNING id::text`,
		req.Nombre, req.Whatsapp, req.Fecha, req.Hora, req.Personas,
		req.Ubicacion, req.Motivo, req.MotivoDetalle, duracionPorMotivo(req.Motivo),
	).Scan(&id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo registrar tu reserva, intenta de nuevo")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"ok": true,
		"id": id,
		"mensaje": "¡Reserva recibida! 💜 Te contactaremos por WhatsApp para confirmarla. " +
			"Recuerda que tienes 15 minutos de tolerancia desde tu hora de reserva; " +
			"pasado ese tiempo, si no te comunicas con nosotros, la mesa puede liberarse.",
	})
}

// ---------------------------------------------------------------------------
// Admin: login + middleware
// ---------------------------------------------------------------------------

func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}
	if !strings.EqualFold(strings.TrimSpace(req.Email), h.Cfg.AdminEmail) ||
		bcrypt.CompareHashAndPassword(h.AdminPassHash, []byte(req.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}
	token, err := auth.Generate(h.Cfg.JWTSecret, h.Cfg.AdminEmail, 12*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo iniciar sesión")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "token": token})
}

func (h *Handler) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			writeError(w, http.StatusUnauthorized, "Falta autenticación")
			return
		}
		sub, err := auth.Verify(h.Cfg.JWTSecret, parts[1])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Sesión inválida o expirada")
			return
		}
		ctx := context.WithValue(r.Context(), "admin", sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ---------------------------------------------------------------------------
// Admin: reservas
// ---------------------------------------------------------------------------

type Reservation struct {
	ID            string `json:"id"`
	Nombre        string `json:"nombre"`
	Whatsapp      string `json:"whatsapp"`
	Fecha         string `json:"fecha"`
	Hora          string `json:"hora"`
	Personas      int    `json:"personas"`
	DuracionMin   int    `json:"duracion_min"`
	Ubicacion     string `json:"ubicacion"`
	Motivo        string `json:"motivo"`
	MotivoDetalle string `json:"motivo_detalle,omitempty"`
	Estado        string `json:"estado"`
	CreatedAt     string `json:"created_at"`
}

func (h *Handler) AdminListReservations(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(
		`SELECT id::text, nombre, whatsapp, fecha::text, hora::text, personas, duracion_min, ubicacion, motivo,
		        COALESCE(motivo_detalle,''), estado, created_at
		 FROM reservations ORDER BY fecha DESC, hora DESC LIMIT 500`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudieron cargar las reservas")
		return
	}
	defer rows.Close()
	list := []Reservation{}
	for rows.Next() {
		var x Reservation
		var created time.Time
		if err := rows.Scan(&x.ID, &x.Nombre, &x.Whatsapp, &x.Fecha, &x.Hora, &x.Personas, &x.DuracionMin,
			&x.Ubicacion, &x.Motivo, &x.MotivoDetalle, &x.Estado, &created); err != nil {
			writeError(w, http.StatusInternalServerError, "Error leyendo reservas")
			return
		}
		x.CreatedAt = created.Format(time.RFC3339)
		list = append(list, x)
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) AdminUpdateReservation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Estado string `json:"estado"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}
	validos := map[string]bool{"pendiente": true, "confirmada": true, "cancelada": true, "liberada": true}
	if !validos[req.Estado] {
		writeError(w, http.StatusUnprocessableEntity, "Estado inválido")
		return
	}
	res, err := h.DB.Exec(`UPDATE reservations SET estado = $1 WHERE id = $2`, req.Estado, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo actualizar")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "Reserva no encontrada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

const defaultWhatsAppTemplate = "Hola {nombre}, te escribimos de Inspirarte Espacio para confirmar tu reserva del {fecha} a las {hora} para {personas} personas en {ubicacion}. Motivo: {motivo}. ¿Nos confirmas tu asistencia? 💜"

var whatsappTemplateFields = map[string]bool{
	"nombre": true, "whatsapp": true, "fecha": true, "hora": true, "duracion": true,
	"personas": true, "ubicacion": true, "motivo": true,
	"motivo_detalle": true, "estado": true, "id": true, "creada_el": true,
}

func validateWhatsAppTemplate(template string) string {
	if template == "" || len([]rune(template)) > 2000 {
		return "Escribe un mensaje de hasta 2000 caracteres"
	}
	for rest := template; rest != ""; {
		open := strings.IndexByte(rest, '{')
		close := strings.IndexByte(rest, '}')
		if close >= 0 && (open < 0 || close < open) {
			return "Hay una llave de cierre sin variable"
		}
		if open < 0 {
			break
		}
		if close < 0 {
			return "Hay una variable sin cerrar"
		}
		field := rest[open+1 : close]
		if !whatsappTemplateFields[field] {
			return "Variable no disponible: {" + field + "}"
		}
		rest = rest[close+1:]
	}
	return ""
}

func (h *Handler) AdminGetWhatsAppTemplate(w http.ResponseWriter, r *http.Request) {
	var template string
	err := h.DB.QueryRow(`SELECT value FROM admin_settings WHERE key = 'whatsapp_reservation_template'`).Scan(&template)
	if err == sql.ErrNoRows {
		template = defaultWhatsAppTemplate
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo cargar el mensaje de WhatsApp")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"template": template})
}

func (h *Handler) AdminSaveWhatsAppTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Template string `json:"template"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Mensaje inválido o demasiado largo")
		return
	}
	req.Template = strings.TrimSpace(req.Template)
	if validationError := validateWhatsAppTemplate(req.Template); validationError != "" {
		writeError(w, http.StatusUnprocessableEntity, validationError)
		return
	}
	if _, err := h.DB.Exec(`INSERT INTO admin_settings (key, value) VALUES ('whatsapp_reservation_template', $1)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`, req.Template); err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo guardar el mensaje")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"template": req.Template})
}

// ---------------------------------------------------------------------------
// Admin: actividades (CRUD)
// ---------------------------------------------------------------------------

func (h *Handler) AdminListActivities(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT ` + activityCols + `, activa FROM activities ORDER BY fecha_hora DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudieron cargar las actividades")
		return
	}
	defer rows.Close()
	acts := []Activity{}
	for rows.Next() {
		var a Activity
		var fechaHora time.Time
		var descripcion, imagen sql.NullString
		if err := rows.Scan(&a.ID, &a.Titulo, &descripcion, &imagen, &fechaHora, &a.Precio,
			&a.CuposTotales, &a.CuposDisponibles, &a.Activa); err != nil {
			writeError(w, http.StatusInternalServerError, "Error leyendo actividades")
			return
		}
		a.Descripcion = descripcion.String
		a.ImagenURL = imagen.String
		a.FechaHora = fechaHora.Format(time.RFC3339)
		a.Agotado = a.CuposDisponibles <= 0
		acts = append(acts, a)
	}
	writeJSON(w, http.StatusOK, acts)
}

type ActivityInput struct {
	Titulo       string  `json:"titulo"`
	Descripcion  string  `json:"descripcion"`
	ImagenURL    string  `json:"imagen_url"`
	FechaHora    string  `json:"fecha_hora"` // RFC3339
	Precio       float64 `json:"precio"`
	CuposTotales int     `json:"cupos_totales"`
	Activa       bool    `json:"activa"`
}

func (h *Handler) AdminCreateActivity(w http.ResponseWriter, r *http.Request) {
	var in ActivityInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}
	in.Titulo = strings.TrimSpace(in.Titulo)
	if in.Titulo == "" || in.CuposTotales < 1 {
		writeError(w, http.StatusUnprocessableEntity, "Título y cupos son obligatorios")
		return
	}
	fechaHora, err := time.Parse(time.RFC3339, in.FechaHora)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Fecha y hora inválidas")
		return
	}
	var id string
	err = h.DB.QueryRow(
		`INSERT INTO activities (titulo, descripcion, imagen_url, fecha_hora, precio, cupos_totales, cupos_disponibles, activa)
		 VALUES ($1, NULLIF($2,''), NULLIF($3,''), $4, $5, $6, $6, $7) RETURNING id::text`,
		in.Titulo, in.Descripcion, in.ImagenURL, fechaHora, in.Precio, in.CuposTotales, in.Activa,
	).Scan(&id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo crear la actividad")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "id": id})
}

func (h *Handler) AdminUpdateActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in ActivityInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}
	in.Titulo = strings.TrimSpace(in.Titulo)
	if in.Titulo == "" || in.CuposTotales < 1 {
		writeError(w, http.StatusUnprocessableEntity, "Título y cupos son obligatorios")
		return
	}
	fechaHora, err := time.Parse(time.RFC3339, in.FechaHora)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Fecha y hora inválidas")
		return
	}
	res, err := h.DB.Exec(
		`UPDATE activities SET titulo=$1, descripcion=NULLIF($2,''), imagen_url=NULLIF($3,''), fecha_hora=$4,
		 precio=$5, cupos_totales=$6, activa=$7,
		 cupos_disponibles = LEAST(cupos_disponibles, $6)
		 WHERE id=$8`,
		in.Titulo, in.Descripcion, in.ImagenURL, fechaHora, in.Precio, in.CuposTotales, in.Activa, id,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo actualizar la actividad")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "Actividad no encontrada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// AdminDeleteActivity retira la actividad de la web (soft delete).
func (h *Handler) AdminDeleteActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := h.DB.Exec(`UPDATE activities SET activa = FALSE WHERE id = $1`, id); err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo retirar la actividad")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
