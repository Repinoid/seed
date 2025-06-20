package handlers

import (
	"fmt"
	"net/http"
)

func Cap(rwr http.ResponseWriter, req *http.Request) {
	rwr.WriteHeader(http.StatusOK)
	fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
}
