package environment

import (
	"context"
	"github.com/zcvaters/gmap-to-gpx/cmd/data"
	"golang.org/x/exp/slog"
	"log"
	"os"
	"strconv"
)

type Environment struct {
	ElevationAPIKey string
	Address         string
	Production      bool
	GCP             *data.GCP
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
			log.Fatal("failed to parse PRODUCTION boolean value")
		}
		e.Production = isProd
	}

	gcpCtx := context.Background()
	elevationApiKey, ok := os.LookupEnv("ELEVATION_API_KEY")
	if !ok || elevationApiKey == "" {
		log.Fatalf("elevation api key not set in enviroinment, have \"%s\"", elevationApiKey)
	}

	if address, ok := os.LookupEnv("ADDRESS"); !ok || address == "" {
		log.Fatalf("missing required environment variable ADDRESS. e.g. localhost:8080, have \"%s\"", address)
	} else {
		e.Address = address
	}

	bucketID, ok := os.LookupEnv("BUCKET_ID")
	if !ok {
		log.Fatal("missing required environment variable BUCKET_ID")
	}
	accessID, ok := os.LookupEnv("ACCESS_ID")
	if !ok {
		log.Fatalf("missing required environment variable ACCESS_ID")
	}
	mapsClient, err := data.NewMapsClient(elevationApiKey)
	if err != nil {
		log.Fatalf("failed to configure maps client: %s", err)
	}

	storageClient, err := data.NewStorageClient(gcpCtx)
	if err != nil {
		log.Fatalf("failed to configure storage client: %s", err)
	}
	gcp := &data.GCP{
		StorageClient: storageClient,
		MapsClient:    mapsClient,
		AccessID:      accessID,
		BucketID:      bucketID,
	}
	e.GCP = gcp
	return e
}
