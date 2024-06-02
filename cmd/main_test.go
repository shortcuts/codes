package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoutes(t *testing.T) {
	t.Setenv("TEST", "foo")

	router := newRouter()

	for _, route := range router.Routes() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, route.Path, nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	}
}
