package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"

	"inspirarte-backend/internal/config"
	"inspirarte-backend/internal/db"
	"inspirarte-backend/internal/handlers"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("No se pudo conectar a Postgres: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("Error en migraciones: %v", err)
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Error preparando credencial admin: %v", err)
	}

	h := &handlers.Handler{DB: database, Cfg: cfg, AdminPassHash: passHash}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors)

	r.Get("/health", h.Health)
	r.Get("/media/activities/{file}", h.ActivityImage)
	r.Route("/api", func(r chi.Router) {
		// Público
		r.Get("/activities", h.ListActivities)
		r.Get("/gallery", h.ListGallery)
		r.Post("/activities/{id}/book", h.BookActivity)
		r.Post("/reservations", h.CreateReservation)

		// Administración
		r.Post("/admin/login", h.AdminLogin)
		r.Group(func(r chi.Router) {
			r.Use(h.RequireAdmin)
			r.Get("/admin/reservations", h.AdminListReservations)
			r.Patch("/admin/reservations/{id}", h.AdminUpdateReservation)
			r.Get("/admin/activity-bookings", h.AdminListActivityBookings)
			r.Patch("/admin/activity-bookings/{id}", h.AdminUpdateActivityBooking)
			r.Get("/admin/whatsapp-template", h.AdminGetWhatsAppTemplate)
			r.Put("/admin/whatsapp-template", h.AdminSaveWhatsAppTemplate)
			r.Get("/admin/gallery", h.AdminListGallery)
			r.Post("/admin/gallery", h.AdminCreateGallery)
			r.Patch("/admin/gallery/{id}", h.AdminUpdateGallery)
			r.Delete("/admin/gallery/{id}", h.AdminDeleteGallery)
			r.Post("/admin/gallery-images", h.AdminUploadGalleryImage)
			r.Get("/admin/activities", h.AdminListActivities)
			r.Post("/admin/activities", h.AdminCreateActivity)
			r.Post("/admin/activity-images", h.AdminUploadActivityImage)
			r.Patch("/admin/activities/{id}", h.AdminUpdateActivity)
			r.Delete("/admin/activities/{id}", h.AdminDeleteActivity)
		})
	})

	log.Printf("API Inspirarte escuchando en :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
