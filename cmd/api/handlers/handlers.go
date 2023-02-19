package handlers

import (
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/environment"
	"go.uber.org/zap"
)

type Handlers struct {
	Environment *environment.Environment
	Log         *zap.SugaredLogger
}
