package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A configuration carried over from the standalone Dev Environments server sets
// BITRISE_MAIN_API_BASE_URL; starting with it would point BITRISE_API_BASE_URL
// (the main API here) at the Dev Environments backend, so startup refuses it.
func TestRunRejectsStandaloneDevEnvironmentsConfig(t *testing.T) {
	t.Setenv("BITRISE_MAIN_API_BASE_URL", "https://api.bitrise.io/v0.1")
	t.Setenv("BITRISE_API_BASE_URL", "https://codespaces-api.services.bitrise.io")

	err := run()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "BITRISE_MAIN_API_BASE_URL is not supported")
		assert.Contains(t, err.Error(), "BITRISE_DEVENV_API_BASE_URL")
	}
}
