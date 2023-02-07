package api

import (
	"github.com/zcvaters/gmap-to-gpx/api/handlers"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/environment"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/router"
	"go.uber.org/zap"
	"net/http"
)

func StartAPI() {
	s := router.CreateNewServer()
	h := &handlers.Handlers{}
	h.Env = environment.CreateNewEnv()
	s.MountHandlers(h)
	logging.Log.Infow("starting API", zap.String("address", h.Env.Address))
	logging.Log.Fatalw("failed to start API", zap.Error(http.ListenAndServe(h.Env.Address, s.Router)))
}
