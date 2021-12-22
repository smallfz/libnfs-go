// Backend interface. To build a customized NFS server you need to implement these.
package fs

import (
	"io"
	"os"
	"time"
)

type FileInfo interface {
	os.FileInfo
	ATime() time.Time
	CTime() time.Time
	NumLinks() int
}

type File interface {
	Name() string
	Stat() (FileInfo, error)
	io.ReadWriteCloser
	io.Seeker
	Truncate() error
	Sync() error
	Readdir(int) ([]FileInfo, error)
}

type FS interface {
	Open(string) (File, error)
	OpenFile(string, int, os.FileMode) (File, error)
	Stat(string) (FileInfo, error)
	Chmod(string, os.FileMode) error
	Rename(string, string) error
	Remove(string) error
	MkdirAll(string, os.FileMode) error
	GetHandle(string) ([]byte, error)
	GetFileId(string) uint64
}

type AllowLink interface {
	Lstat(string) (FileInfo, error)
	Symlink(string, string) error
}
