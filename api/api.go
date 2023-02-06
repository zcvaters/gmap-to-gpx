package api

import (
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/environment"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/router"
	"go.uber.org/zap"
	"net/http"
)

func StartAPI() {
	environment.Initialize()
	s := router.CreateNewServer()
	s.MountHandlers()
	logging.Log.Infow("starting API", zap.String("address", environment.Variables.Address))
	logging.Log.Fatalw("failed to start API", zap.Error(http.ListenAndServe(environment.Variables.Address, s.Router)))
}
