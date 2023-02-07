package test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/zcvaters/gmap-to-gpx/api/handlers"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/environment"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/router"
	"github.com/zcvaters/gmap-to-gpx/cmd/data"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConvertGMAPToGPX(t *testing.T) {
	type test struct {
		input  any
		status int
		body   *data.ResponseData
	}

	environment.Initialize()
	s := router.CreateNewServer()
	s.MountHandlers()
	tt := []test{
		{input: nil, status: http.StatusBadRequest, body: &data.ResponseData{Data: "", Error: "request json malformed"}},
		{input: `{"test": 123}`, status: http.StatusBadRequest, body: &data.ResponseData{Data: "", Error: "request json malformed"}},
		{input: &handlers.GMapToGPXRequest{RouteID: 4999999}, status: http.StatusBadRequest, body: &data.ResponseData{Data: "", Error: "invalid route ID, must be greater than 5000000."}},
		{input: &handlers.GMapToGPXRequest{RouteID: 5000001}, status: http.StatusOK},
		{input: &handlers.GMapToGPXRequest{RouteID: 7696696}, status: http.StatusOK},
	}
	for _, tc := range tt {

		reqBody, err := json.Marshal(tc.input)
		if err != nil {
			t.Errorf("failed to marshal request data: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, "/api/v1/gMapToGPX", bytes.NewReader(reqBody))
		if err != nil {
			t.Errorf("failed to create a new request: %v", err)
		}
		rr := executeRequest(req, s)
		checkResponseCode(t, tc.status, rr.Code)
		if tc.body != nil {
			j, _ := json.Marshal(tc.body)
			tcJsonStr := string(j)
			require.JSONEq(t, tcJsonStr, rr.Body.String())
		}
	}
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func executeRequest(req *http.Request, s *router.Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(rr, req)

	return rr
}
