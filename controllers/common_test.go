//go:build types
// +build types

package controllers

import (
	"os"
	"testing"
)

func TestGetenv_PROJECT(t *testing.T) {
	oldProject := Project
	defer func() {
		Project = oldProject
	}()

	os.Setenv("PROJECT", "")

	_, err := getenv("PROJECT")
	if err == nil {
		t.Error("expected error")
	}
}
