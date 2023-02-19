package data

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/joomcode/errorx"
	"googlemaps.github.io/maps"
	"io"
	"net/http"
	"time"
)

type GCP struct {
	StorageClient *storage.Client
	MapsClient    *maps.Client
	AccessID      string
	BucketID      string
}

func NewStorageClient(ctx context.Context) (*storage.Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to create storage client")
	}

	return client, nil
}

func NewMapsClient(apiKey string) (*maps.Client, error) {
	gClient, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to configure maps api client: %q", err)
	}

	return gClient, nil
}

// GetSignedUploadURL fetches a PUT url for an object.
func (g *GCP) GetSignedUploadURL(obj string) (*string, error) {
	url, err := g.StorageClient.Bucket(g.BucketID).SignedURL(obj, &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         http.MethodPut,
		Expires:        time.Now().Add(15 * time.Minute),
		GoogleAccessID: g.AccessID,
		ContentType:    "binary/octet-stream",
	})

	if err != nil {
		return nil, errorx.Decorate(err, "failed to get upload URL. Bucket: %s", g.BucketID)
	}

	return &url, nil
}

// GetSignedDownloadURL fetches the given object string to a pre signed download URL.
func (g *GCP) GetSignedDownloadURL(obj string) (*string, error) {
	url, err := g.StorageClient.Bucket(g.BucketID).SignedURL(obj, &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         http.MethodGet,
		Expires:        time.Now().Add(15 * time.Minute),
		GoogleAccessID: g.AccessID,
	})

	if err != nil {
		return nil, errorx.Decorate(err, "failed to get signed download URL. Bucket: %s ", g.BucketID)
	}
	return &url, nil
}

func (g *GCP) StreamFileUpload(parent context.Context, object string, payload []byte) error {
	buf := bytes.NewBuffer(payload)

	ctx, cancel := context.WithTimeout(parent, time.Second*50)
	defer cancel()

	wc := g.StorageClient.Bucket(g.BucketID).Object(object).NewWriter(ctx)
	wc.ChunkSize = 0

	if _, err := io.Copy(wc, buf); err != nil {
		return errorx.Decorate(err, "failed to copy buffer")
	}
	if err := wc.Close(); err != nil {
		return errorx.Decorate(err, "failed to close Writer.Closer")
	}

	return nil
}
