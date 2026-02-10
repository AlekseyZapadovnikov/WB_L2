package transport

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter инициализирует chi роутер и регистрирует хендлеры
func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(loggingMiddleware)

	r.Post("/create_event", h.CreateEvent)
	r.Post("/update_event", h.UpdateEvent)
	r.Post("/delete_event", h.DeleteEvent)

	r.Get("/events_for_day", h.EventsForDay)
	r.Get("/events_for_week", h.EventsForWeek)
	r.Get("/events_for_month", h.EventsForMonth)

	return r
}

// loggingMiddleware логирует запросы (адаптация под http.Handler)
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		
		next.ServeHTTP(ww, r)

		log.Printf(
			"[%s] %s %s | Status: %d | Size: %d | Duration: %s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			ww.Status(),
			ww.BytesWritten(),
			time.Since(start),
		)
	})
}