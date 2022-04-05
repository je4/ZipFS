package ZipFS

import (
	"os"
	"testing"
)

func TestOpen(t *testing.T) {
	zfs, err := NewZipFS(os.DirFS("."), "testdata/test.zip")
	if err != nil {
		t.Fatalf("cannot open test zip: %v", err)
	}
	defer zfs.Close()

	//sfs, err := fs.Sub(zfs, "zip")

	fname := "zip/subfolder1/subfolder11/test02 - Kopie.txt"
	f, err := zfs.Open(fname)
	if err != nil {
		t.Fatalf("cannot file %s in zip: %v", fname, err)
	}
	defer f.Close()
}
