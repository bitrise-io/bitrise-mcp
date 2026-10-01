package devenvironments

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func tarGz(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for name, body := range entries {
		assert.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}))
		_, _ = tw.Write([]byte(body))
	}
	assert.NoError(t, tw.Close())
	assert.NoError(t, gw.Close())
	return buf.Bytes()
}

func TestExtractTarGzRejectsEscapes(t *testing.T) {
	t.Run("parent traversal", func(t *testing.T) {
		dest := t.TempDir()
		err := extractTarGz(bytes.NewReader(tarGz(t, map[string]string{"../evil": "x"})), dest)
		assert.Error(t, err)
		_, statErr := os.Stat(filepath.Join(filepath.Dir(dest), "evil"))
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("symlinked parent directory", func(t *testing.T) {
		outside := t.TempDir()
		dest := t.TempDir()
		assert.NoError(t, os.Symlink(outside, filepath.Join(dest, "sub")))
		err := extractTarGz(bytes.NewReader(tarGz(t, map[string]string{"sub/file": "x"})), dest)
		assert.Error(t, err)
		_, statErr := os.Stat(filepath.Join(outside, "file"))
		assert.True(t, os.IsNotExist(statErr), "must not write through the symlinked parent")
	})

	t.Run("regular entries extract", func(t *testing.T) {
		dest := t.TempDir()
		assert.NoError(t, extractTarGz(bytes.NewReader(tarGz(t, map[string]string{"a/b.txt": "hello"})), dest))
		got, err := os.ReadFile(filepath.Join(dest, "a", "b.txt"))
		assert.NoError(t, err)
		assert.Equal(t, "hello", string(got))
	})
}

// The archive is spooled to a temporary file before extraction, so its size is
// bounded by disk rather than memory, and the file is readable from the start.
func TestDownloadArchiveSpoolsToTempFile(t *testing.T) {
	archive := tarGz(t, map[string]string{"out/app.ipa": "binary"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	f, err := downloadArchive(context.Background(), srv.URL)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	dest := t.TempDir()
	assert.NoError(t, extractTarGz(f, dest))
	got, err := os.ReadFile(filepath.Join(dest, "out", "app.ipa"))
	assert.NoError(t, err)
	assert.Equal(t, "binary", string(got))
}
