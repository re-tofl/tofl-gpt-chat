package utils

import (
	"encoding/json"
	"net/http"
)

func DecodeBody(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}
