package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/zcvaters/gmap-to-gpx/cmd/api/handlers"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"time"
)

type Handler func(w http.ResponseWriter, r *http.Request) http.Handler

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler := h(w, r); handler != nil {
		handler.ServeHTTP(w, r)
	}
}

func (s *Server) MountHandlers(h *handlers.Handlers) {
	slogJSONHandler := slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}.NewJSONHandler(os.Stdout)
	s.Router.Use(middleware.RequestID)
	s.Router.Use(logging.NewStructuredLogger(slogJSONHandler))
	s.Router.Use(middleware.AllowContentType("application/json"))
	s.Router.Use(middleware.Heartbeat("/"))
	s.Router.Use(middleware.RealIP)
	s.Router.Use(httprate.LimitByIP(10, 1*time.Minute))
	s.Router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"},
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	s.Router.Use(middleware.Recoverer)

	s.Router.Route("/api/v1", func(r chi.Router) {
		r.Method("POST", "/gMapToGPX", Handler(h.ConvertGMAPToGPX))
	})
}

type Server struct {
	Router *chi.Mux
}

func CreateNewServer() *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	return s
}
