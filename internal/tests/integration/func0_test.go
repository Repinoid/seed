package tests

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCap(t *testing.T) {
	port := os.Getenv("PORT")
//	hoster, exists := os.LookupEnv("ADDRESS")

	resp, err := http.Get("http://0.0.0.0:" + port + "/cap")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	_ = body
}
