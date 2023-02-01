package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/zcvaters/gmap-to-gpx/api/configure/environment"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/mux"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	environment.Initialize()
	logging.Initialize(environment.Variables.Production)
	m := chi.NewRouter()
	s := mux.Mux{Router: m}
	s.InitializeRouter()
	logging.Log.Infow("starting API", zap.String("address", environment.Variables.Address))
	logging.Log.Fatalw("failed to start API", zap.Error(http.ListenAndServe(environment.Variables.Address, m)))
}
