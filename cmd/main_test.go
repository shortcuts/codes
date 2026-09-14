package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoutes(t *testing.T) {
	s := newServer()
	defer s.close()

	for _, route := range routes {
		if err := s.registerRoute(route); err != nil {
			t.Fatalf("register %s: %v", route.path, err)
		}
	}

	for _, route := range routes {
		req := httptest.NewRequest(http.MethodGet, "/"+route.path, nil)
		rec := httptest.NewRecorder()

		s.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("GET /%s: got status %d, want %d", route.path, rec.Code, http.StatusOK)
		}
	}
}

func TestNotFound(t *testing.T) {
	s := newServer()
	defer s.close()

	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
	}
}
