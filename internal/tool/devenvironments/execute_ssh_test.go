package devenvironments

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDNSFailureHint(t *testing.T) {
	dnsErr := &net.DNSError{Err: "no such host", Name: "foo.bar", IsNotFound: true}

	t.Run("darwin gets the networksetup command", func(t *testing.T) {
		hint := dnsFailureHint(dnsErr, "darwin")
		assert.NotEmpty(t, hint)
		assert.Contains(t, hint, "networksetup -setdnsservers")
	})

	t.Run("non-darwin gets a generic hint without the mac-specific command", func(t *testing.T) {
		hint := dnsFailureHint(dnsErr, "linux")
		assert.NotEmpty(t, hint)
		assert.NotContains(t, hint, "networksetup -setdnsservers")
	})

	t.Run("non-DNS error yields no hint", func(t *testing.T) {
		hint := dnsFailureHint(errors.New("connection refused"), "darwin")
		assert.Empty(t, hint)
	})

	t.Run("DNS error wrapped as dialSSH wraps it is still detected", func(t *testing.T) {
		wrapped := fmt.Errorf("ssh dial %s: %w", "foo.bar:22", dnsErr)
		hint := dnsFailureHint(wrapped, "darwin")
		assert.NotEmpty(t, hint)
	})

	t.Run("dial timeout (successful lookup, unreachable resolved IP) gets the timeout-specific hint", func(t *testing.T) {
		wrapped := fmt.Errorf("ssh dial %s: %w", "foo.bar:22", fakeNetErr{timeout: true})
		hint := dnsFailureHint(wrapped, "darwin")
		assert.NotEmpty(t, hint)
		assert.Contains(t, hint, "dial timed out")
		assert.Contains(t, hint, "networksetup -setdnsservers")
	})

	t.Run("non-timeout net.Error yields no hint", func(t *testing.T) {
		wrapped := fmt.Errorf("ssh dial %s: %w", "foo.bar:22", fakeNetErr{timeout: false})
		hint := dnsFailureHint(wrapped, "darwin")
		assert.Empty(t, hint)
	})
}

// fakeNetErr is a minimal net.Error for exercising the timeout branch of
// dnsFailureHint without depending on a real dial actually timing out.
type fakeNetErr struct{ timeout bool }

func (e fakeNetErr) Error() string   { return "fake net error" }
func (e fakeNetErr) Timeout() bool   { return e.timeout }
func (e fakeNetErr) Temporary() bool { return e.timeout }
