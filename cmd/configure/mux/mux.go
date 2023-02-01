package mux

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joomcode/errorx"
	"github.com/zcvaters/gmap-to-gpx/api/configure"
	"github.com/zcvaters/gmap-to-gpx/api/handlers"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	. "github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	"github.com/zcvaters/gmap-to-gpx/cmd/data"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
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

		if err := data.WriteJSONBytes(configure.ResponseData{
			Data:  make([]any, 0),
			Error: errX.Message(),
		}, w); err != nil {
			Log.Error(errorx.Decorate(err, "failed to write JSON response: %v"))
		}
	}
}

type Mux struct {
	Router *chi.Mux
}

func (m *Mux) InitializeRouter() {
	slogJSONHandler := slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}.NewJSONHandler(os.Stdout)
	m.Router.Use(middleware.RequestID)
	m.Router.Use(logging.NewStructuredLogger(slogJSONHandler))
	m.Router.Use(middleware.Recoverer)
	m.Router.Method("POST", "/gMapToGPX", Handler(handlers.ConvertGMAPToGPX))
}
