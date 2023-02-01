package data

import (
	"encoding/json"
	"net/http"
)

func MarshalJSON(v any) (*[]byte, error) {
	r, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func WriteJSONBytes(v any, w http.ResponseWriter) error {
	r, err := MarshalJSON(v)
	if err != nil {
		return err
	}

	_, err = w.Write(*r)
	if err != nil {
		return err
	}

	return nil
}
