package unixfs

import (
	"os"

	"github.com/smallfz/libnfs-go/fs"
)

type file struct {
	os.File
}

func (f *file) Readdir(n int) ([]fs.FileInfo, error) {
	infos, err := f.File.Readdir(n)
	if err != nil {
		return nil, err
	}

	var fis []fs.FileInfo
	for _, info := range infos {
		fis = append(fis, statFromInfo(info))
	}

	return fis, nil
}

func (f *file) Stat() (fs.FileInfo, error) {
	info, err := f.File.Stat()
	if err != nil {
		return nil, err
	}

	return statFromInfo(info), nil
}

func (f *file) Truncate() error {
	return f.File.Truncate(0)
}
