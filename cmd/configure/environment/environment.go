package environment

import (
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	"golang.org/x/exp/slog"
	"os"
	"strconv"
)

type Environment struct {
	ElevationAPIKey string
	Address         string
	Production      bool
}

func CreateNewEnv() *Environment {
	e := &Environment{}
	isProd := false
	prodVal, ok := os.LookupEnv("PRODUCTION")
	if !ok || prodVal == "" {
		slog.Info("failed to get PRODUCTION environment variable, defaulting to development")
	} else {
		var err error
		if isProd, err = strconv.ParseBool(prodVal); err != nil {
			slog.Info("failed to parse PRODUCTION boolean value")
			os.Exit(1)
		}
		e.Production = isProd
	}
	logging.Log = logging.CreateNewLogger(e.Production)

	if elevationApiKey, ok := os.LookupEnv("ELEVATION_API_KEY"); !ok || elevationApiKey == "" {
		logging.Log.Fatalf("elevation api key not set in enviroinment, have \"%s\"", elevationApiKey)
	} else {
		e.ElevationAPIKey = elevationApiKey
	}

	if address, ok := os.LookupEnv("ADDRESS"); !ok || address == "" {
		logging.Log.Fatalf("missing required environment variable ADDRESS. e.g. localhost:8080, have \"%s\"", address)
	} else {
		e.Address = address
	}

	return e
}
