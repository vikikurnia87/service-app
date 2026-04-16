package handler_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"service-app/pkg/response"
)

// parseResp parses the JSON response body into the response envelope.
func parseResp(t *testing.T, rec *httptest.ResponseRecorder) response.Response {
	t.Helper()
	var resp response.Response
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err, "failed to parse response body")
	return resp
}
