package ZipFS

import "io/fs"

type DirEntry struct {
	FileInfo
}

func (de *DirEntry) Name() string {
	return de.FileInfo.Name()
}

func (de *DirEntry) IsDir() bool {
	return de.FileInfo.IsDir()
}

func (de *DirEntry) Type() fs.FileMode {
	return de.FileInfo.Mode()
}

func (de *DirEntry) Info() (fs.FileInfo, error) {
	return fs.FileInfo(de.FileInfo), nil
}
