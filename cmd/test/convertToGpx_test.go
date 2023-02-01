package test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConvertGMAPToGPX(t *testing.T) {
	type test struct {
		input  *handlers.GMapToGPXRequest
		status int
	}

	r := mux.NewRouter()
	mux.InitializeRouter(r)
	tt := []test{
		{input: nil, status: http.StatusBadRequest},
		{input: &handlers.GMapToGPXRequest{RouteID: 5000001}, status: http.StatusOK},
		//{input: &handlers.GMapToGPXRequest{RouteID: 4999999}, status: http.StatusBadRequest},
		//{input: &handlers.GMapToGPXRequest{RouteID: 7696696}, status: http.StatusOK},
	}
	for _, tc := range tt {

		reqBody, err := json.Marshal(tc.input)
		if err != nil {
			t.Errorf("failed to marshal request data: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, "/api/gMapToGPX", bytes.NewReader(reqBody))
		if err != nil {
			t.Errorf("failed to create a new request: %v", err)
		}
		w := httptest.NewRecorder()
		err = r.ServeHTTPError(w, req)
		//r.ServeHTTP(rr, req)
		require.NoError(t, err)

		//if status := rr.Code; status != tc.status {
		//	t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d. RouteID: %v", http.StatusOK, status, tc.input.RouteID)
		//}
	}
}
