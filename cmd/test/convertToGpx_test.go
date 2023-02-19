package test

import (
	"bytes"
	"encoding/json"
	"github.com/zcvaters/gmap-to-gpx/cmd/api/handlers"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/environment"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/router"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var server = router.CreateNewServer()

func TestMain(m *testing.M) {
	h := &handlers.Handlers{}
	h.Environment = environment.CreateNewEnv()
	server.MountHandlers(h)
	os.Exit(m.Run())
}

func TestConvertGMAPToGPX(t *testing.T) {
	type test struct {
		desc   string
		input  any
		status int
	}

	tt := []test{
		{desc: "Nil input", input: nil, status: http.StatusBadRequest},
		{desc: "Invalid json to unmarshal", input: `{"test": 123}`, status: http.StatusBadRequest},
		{desc: "Invalid Route ID", input: &handlers.GMapToGPXRequest{RouteID: 4999999}, status: http.StatusBadRequest},
		{desc: "Valid Route ID", input: &handlers.GMapToGPXRequest{RouteID: 5000001}, status: http.StatusOK},
		{desc: "Random Route ID", input: &handlers.GMapToGPXRequest{RouteID: 7696696}, status: http.StatusOK},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.input)
			if err != nil {
				t.Errorf("failed to marshal request data: %v", err)
			}

			req, err := http.NewRequest(http.MethodPost, "/api/v1/gMapToGPX", bytes.NewReader(reqBody))
			if err != nil {
				t.Errorf("failed to create a new request: %v", err)
			}
			rr := executeRequest(req, server)
			checkResponseCode(t, tc.status, rr.Code)
			t.Logf("Body: %q", rr.Body)
		})
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
