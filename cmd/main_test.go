package main

import (
	"testing"
)

func TestRoutes(t *testing.T) {
	t.Setenv("TEST", "foo")
	t.Skip()
}
