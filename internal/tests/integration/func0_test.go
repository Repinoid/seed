package tests

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCap(t *testing.T) {
	resp, err := http.Get("http://0.0.0.0:8100/cap")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	_ = body
}
