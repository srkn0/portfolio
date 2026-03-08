package util

import "net/http"

func IsHxRequest(req *http.Request) bool {
	return req.Header.Get("HX-Request") == "true"
}
