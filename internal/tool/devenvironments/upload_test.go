package devenvironments

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// Symlinks in an uploaded folder keep their target, so they are not recreated
// on the VM as links to "".
func TestCreateTarGzKeepsSymlinkTargets(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "real.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real.txt", filepath.Join(src, "link.txt")); err != nil {
		t.Fatal(err)
	}

	archivePath, err := createTarGz(src)
	if err != nil {
		t.Fatalf("createTarGz: %v", err)
	}
	defer os.Remove(archivePath)

	f, err := os.Open(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			t.Fatal("link.txt not found in the archive")
		}
		if err != nil {
			t.Fatal(err)
		}
		if h.Name != "link.txt" {
			continue
		}
		if h.Typeflag != tar.TypeSymlink || h.Linkname != "real.txt" {
			t.Errorf("link.txt: typeflag %q linkname %q, want symlink to real.txt", h.Typeflag, h.Linkname)
		}
		return
	}
}
