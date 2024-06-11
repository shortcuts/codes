package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoutes(t *testing.T) {
	t.Setenv("TEST", "foo")

	s := newServer()

	for _, route := range routes {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", route), nil)
		s.http.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	}
}
