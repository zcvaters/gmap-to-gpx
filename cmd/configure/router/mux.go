package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/joomcode/errorx"
	"github.com/zcvaters/gmap-to-gpx/api/handlers"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	. "github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	"github.com/zcvaters/gmap-to-gpx/cmd/data"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"time"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		errX := errorx.Cast(err)
		Log.Error(errorx.Decorate(errX, "handler error"))
		if errX.IsOfType(errorx.AssertionFailed) || errX.IsOfType(errorx.IllegalArgument) {
			w.WriteHeader(http.StatusBadRequest)
		} else if errorx.IsOfType(errX, errorx.InternalError) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		if err := data.WriteJSONBytes(data.ResponseData{
			Data:  "",
			Error: errX.Message(),
		}, w); err != nil {
			Log.Error(errorx.Decorate(err, "failed to write JSON response: %v"))
		}
	}
}

func (s *Server) MountHandlers() {
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
	s.Router.Use(middleware.Recoverer)

	s.Router.Route("/api/v1", func(r chi.Router) {
		r.Method("POST", "/gMapToGPX", Handler(handlers.ConvertGMAPToGPX))
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
