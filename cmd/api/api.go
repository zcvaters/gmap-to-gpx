package api

import (
	"github.com/zcvaters/gmap-to-gpx/cmd/api/handlers"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/environment"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/router"
	"go.uber.org/zap"
	"net/http"
)

func StartAPI() {
	s := router.CreateNewServer()
	env := environment.CreateNewEnv()
	log := logging.CreateNewLogger(env.Production)
	h := &handlers.Handlers{
		Environment: env,
		Log:         log,
	}
	s.MountHandlers(h)
	log.Infow("starting API", zap.String("address", env.Address))
	log.Fatalw("failed to start API", zap.Error(http.ListenAndServe(env.Address, s.Router)))
	if err := env.GCP.StorageClient.Close(); err != nil {
		log.Errorf("failed to close GCP client: %v", err)
	}
}
