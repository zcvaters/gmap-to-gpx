package environment

import (
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	"os"
	"strconv"
)

type Environment struct {
	OpenElevationAPIURL string
	Address             string
	Production          bool
}

var Variables Environment

func Initialize() {
	Variables.SetVariables()
}

func (e *Environment) SetVariables() {
	if elevationApiUrl, ok := os.LookupEnv("ELEVATION_API"); !ok || elevationApiUrl == "" {
		logging.Log.Fatalf("elevation api environment variable not set, have \"%s\"", elevationApiUrl)
	} else {
		e.OpenElevationAPIURL = elevationApiUrl
	}

	if address, ok := os.LookupEnv("ADDRESS"); !ok || address == "" {
		logging.Log.Fatalf("missing required environment variable ADDRESS. e.g. localhost:8080, have \"%s\"", address)
	} else {
		e.Address = address
	}

	isProd := false
	prodVal, ok := os.LookupEnv("PRODUCTION")
	if !ok || prodVal == "" {
		logging.Log.Warn("failed to get PRODUCTION environment variable, defaulting to development")
	} else {
		var err error
		if isProd, err = strconv.ParseBool(prodVal); err != nil {
			logging.Log.Fatal("failed to parse PRODUCTION boolean value")
		}
		e.Production = isProd
	}
}
