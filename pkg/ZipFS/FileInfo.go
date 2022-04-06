package ZipFS

import (
	"io/fs"
	"time"
)

type FileInfo struct {
	zipFS     *ZipFS
	name      string
	size      int64
	isDir     bool
	isRegular bool
	modTime   time.Time
	mode      fs.FileMode
}

func (de FileInfo) Mode() fs.FileMode {
	return de.mode
}

func (de FileInfo) ModTime() time.Time {
	return de.modTime
}

func (de FileInfo) Name() string {
	return de.name
}

func (de FileInfo) Size() int64 {
	return de.size
}

func (de FileInfo) IsDir() bool {
	return de.isDir
}

func (de FileInfo) IsRegular() bool {
	return de.isRegular
}

func (de FileInfo) Type() fs.FileMode {
	return de.mode
}

func (de FileInfo) Sys() interface{} {
	return nil
}
