package unixfs

import (
	"io/fs"
	"os"
)

type FileInfo struct {
	fs.FileInfo
}

func Stat(name string) (*FileInfo, error) {
	fi, err := os.Lstat(name) // Do not follow links
	if err != nil {
		return nil, err
	}

	return statFromInfo(fi), nil
}

func statFromInfo(fi fs.FileInfo) *FileInfo {
	return &FileInfo{
		FileInfo: fi,
	}
}
