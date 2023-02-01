package data

import (
	"github.com/joomcode/errorx"
	"net/http"
)

type Error struct {
	StatusCode int    `json:"-"`
	Code       int    `json:"-"`
	Message    string `json:"message,omitempty"`
	Internal   int    `json:"internal,omitempty"`
}

var NoRequestBody = errorx.CommonErrors.NewType("no_body")

var Errors = map[int]Error{
	1000: {
		Code:       1000,
		Message:    "Unknown internal server error.",
		Internal:   1,
		StatusCode: http.StatusInternalServerError,
	},
	1001: {
		Code:       1001,
		Message:    "Invalid routeID. Route ID must be greater than 4999999.",
		Internal:   2,
		StatusCode: http.StatusBadRequest,
	},
	1002: {
		Code:       1002,
		Message:    "Invalid request contents.",
		Internal:   3,
		StatusCode: http.StatusBadRequest,
	},
	1003: {
		Code:       1003,
		Message:    "Malformed request contents.",
		Internal:   4,
		StatusCode: http.StatusBadRequest,
	},
	1004: {
		Code:       1004,
		Message:    "Elevation data could not be acquired.",
		Internal:   5,
		StatusCode: http.StatusInternalServerError,
	},
}
