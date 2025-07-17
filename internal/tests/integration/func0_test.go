package tests

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func aTestCap(t *testing.T) {

	port, exists := os.LookupEnv("PORTOS")
	if !exists {
		port = "8080"
	}

	resp, err := http.Get("http://0.0.0.0:" + port + "/")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	_ = body
}
