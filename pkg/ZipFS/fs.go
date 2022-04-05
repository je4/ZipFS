package ZipFS

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type ReaderAtCloser interface {
	io.ReaderAt
	io.ReadCloser
}

type ZipFS struct {
	parentFS   fs.FS
	path       string
	zipSize    int64
	parentFile ReaderAtCloser
	zip        *zip.Reader
}

func NewZipFS(fsys fs.FS, path string) (*ZipFS, error) {
	zfs := &ZipFS{
		parentFS: fsys,
		path:     path,
		zip:      nil,
	}

	fi, err := fs.Stat(fsys, path)
	if err != nil {
		return nil, errors.Wrapf(err, "NewZipFS: cannot stat file %s", path)
	}
	if fi.IsDir() {
		return nil, errors.New(fmt.Sprintf("NewZipFS: %s is a directory", path))
	}
	zfs.zipSize = fi.Size()
	return zfs, nil
}

func (zfs *ZipFS) openZip() error {
	if zfs.zip == nil {
		f, err := zfs.parentFS.Open(zfs.path)
		if err != nil {
			return errors.Wrapf(err, "zfs.openZip: cannot open zip file %s", zfs.path)
		}
		var ok bool
		zfs.parentFile, ok = f.(ReaderAtCloser)
		if !ok {
			return errors.New(fmt.Sprintf("zfs.openZip filesystem does not provide io.ReaderAt for %s", zfs.path))
		}
		zfs.zip, err = zip.NewReader(zfs.parentFile, zfs.zipSize)
		if err != nil {
			_ = zfs.parentFile.Close()
			return errors.Wrapf(err, "zfs.openZip: cannot open Zip file from file handle %s", zfs.path)
		}
	}
	return nil
}
func (zfs *ZipFS) Open(path string) (fs.File, error) {
	path = filepath.ToSlash(filepath.Clean(path))
	if err := zfs.openZip(); err != nil {
		return nil, err
	}

	f, err := zfs.zip.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "zfs.Open: cannot open %s inside %s", path, zfs.path)
	}
	return f, nil
}

func (zfs *ZipFS) Close() error {
	zfs.zip = nil
	if err := zfs.parentFile.Close(); err != nil {
		return errors.Wrapf(err, "zfs.Close: cannot close parent file %s", zfs.path)
	}
	return nil
}

func (zfs *ZipFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if err := zfs.openZip(); err != nil {
		return nil, errors.Wrapf(err, "zfs.ReadDir: cannot open zip %s directory %s", zfs.path, name)
	}

	result := []*DirEntry{}
	addIfNew := func(de *DirEntry) {
		found := false
		for _, de := range result {
			if de.Name() == name {
				found = true
				break
			}
		}
		if !found {
			result = append(result, de)
		}
	}

	name = filepath.ToSlash(filepath.Clean(name))

	for _, f := range zfs.zip.File {

	}
}

func fileInfo(f *zip.File, zfs *ZipFS, name string, nameDir string) (fs.FileInfo, error) {
	fi := &FileInfo{
		zipFS:     zfs,
		name:      name,
		size:      0,
		isDir:     false,
		isRegular: false,
		modTime:   time.Time{},
		mode:      0,
	}
	if f.Name == name {
		fi.isRegular = true
		fi.size = int64(f.UncompressedSize64)
		fi.modTime = f.Modified
		fi.mode = f.Mode()
		return fi, nil
	} else if strings.HasPrefix(nameDir, f.Name) {
		fi.isDir = true
		fi.mode = fs.ModeDir | fs.ModePerm
		return fi, nil
	}
	return nil, fs.ErrNotExist
}

func (zfs *ZipFS) fileInfo(name string) (fs.FileInfo, error) {
	nameDir := strings.TrimRight(name, "/") + "/"
	for _, f := range zfs.zip.File {
		fileInfo(f, zfs, name, nameDir)
	}
	return nil, fs.ErrNotExist
}

func (zfs *ZipFS) isDir(name string) (bool, error) {
	name = strings.TrimRight(name, "/") + "/"
	for _, fi := range zfs.zip.File {
		if fi.Name == name {
			return false, nil
		}
		if strings.HasPrefix(fi.Name, name) {
			return true, nil
		}
	}
	return false, fs.ErrNotExist
}

func (zfs *ZipFS) size(name string) (int64, error) {
	fi, err := zfs.fileInfo(name)
	if err != nil {
		return 0, errors.Wrapf(err, "zfs.size: cannot get fileInfo of %s in %s", name, zfs.path)
	}
	return fi.Size(), nil
}
