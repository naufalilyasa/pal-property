package searchindex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_UpsertDocument_UsesPutAndJSONBody(t *testing.T) {
	var method, path string
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "", "", server.Client())
	require.NoError(t, err)

	err = client.UpsertDocument(context.Background(), "listings", "listing-1", map[string]any{"title": "Villa"})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, method)
	assert.Equal(t, "/listings/_doc/listing-1", path)
	assert.Equal(t, "Villa", payload["title"])
}

func TestClient_DeleteDocument_UsesDelete(t *testing.T) {
	var method, path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "", "", server.Client())
	require.NoError(t, err)

	err = client.DeleteDocument(context.Background(), "listings", "listing-1")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, method)
	assert.Equal(t, "/listings/_doc/listing-1", path)
}
