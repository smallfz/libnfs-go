package memfs

import (
	"os"
	"time"
)

type fileInfo struct {
	name     string
	perm     os.FileMode
	size     int64
	modTime  time.Time
	aTime    time.Time
	cTime    time.Time
	isDir    bool
	numLinks int
}

func (fi fileInfo) Name() string {
	return fi.name
}

func (fi fileInfo) Size() int64 {
	return fi.size
}

func (fi fileInfo) Mode() os.FileMode {
	return fi.perm
}

func (fi fileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi fileInfo) ATime() time.Time {
	return fi.aTime
}

func (fi fileInfo) CTime() time.Time {
	return fi.cTime
}

func (fi fileInfo) IsDir() bool {
	return fi.isDir
}

func (fi fileInfo) Sys() interface{} {
	return nil
}

func (fi fileInfo) NumLinks() int {
	return fi.numLinks
}
