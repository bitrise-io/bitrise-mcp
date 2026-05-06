package registration_test

import (
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
)

func TestRegistrationEndpointsAreNotPlaceholders(t *testing.T) {
	if strings.Contains(bitrise.APIRegistrationBaseURL, "placeholder") {
		t.Fatal("APIRegistrationBaseURL is still a placeholder — replace it with the real endpoint before merging")
	}
}
