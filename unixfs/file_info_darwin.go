package unixfs

import (
	"syscall"
	"time"
)

func (fi FileInfo) ATime() time.Time {
	return time.Unix(fi.FileInfo.Sys().(*syscall.Stat_t).Atimespec.Unix())
}

func (fi FileInfo) CTime() time.Time {
	return time.Unix(fi.FileInfo.Sys().(*syscall.Stat_t).Ctimespec.Unix())
}

func (fi FileInfo) NumLinks() int {
	return int(fi.FileInfo.Sys().(*syscall.Stat_t).Nlink)
}

func (fi FileInfo) Inode() uint64 {
	return fi.FileInfo.Sys().(*syscall.Stat_t).Ino
}
