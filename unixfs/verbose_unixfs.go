package unixfs

import (
	"os"

	"github.com/smallfz/libnfs-go/fs"
	"github.com/smallfz/libnfs-go/log"
)

type VerboseUnixFS struct {
	UnixFS
}

func NewVerbose(workdir string) (*VerboseUnixFS, error) {
	fs, err := New(workdir)
	if err != nil {
		return nil, err
	}

	return &VerboseUnixFS{UnixFS: *fs}, nil
}

func (s *VerboseUnixFS) Chmod(name string, mode os.FileMode) error {
	log.Infof("unixfs.Chmod(%s, %o)", name, mode)

	err := s.UnixFS.Chmod(name, mode)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

func (s *VerboseUnixFS) Chown(name string, uid, gid int) error {
	log.Infof("unixfs.Chown(%s, %d, %d)", name, uid, gid)

	err := s.UnixFS.Chown(name, uid, gid)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

func (s *VerboseUnixFS) MkdirAll(name string, mode os.FileMode) error {
	log.Infof("unixfs.MkdirAll(%s, %o)", name, mode)

	err := s.UnixFS.MkdirAll(name, mode)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

func (s *VerboseUnixFS) Remove(name string) error {
	log.Infof("unixfs.Remove(%s)", name)

	err := s.UnixFS.Remove(name)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

func (s *VerboseUnixFS) Rename(oldpath, newpath string) error {
	log.Infof("unixfs.Rename(%s, %s)", oldpath, newpath)

	err := s.UnixFS.Rename(oldpath, newpath)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

func (s *VerboseUnixFS) Link(oldpath, newpath string) error {
	log.Infof("unixfs.Link(%s, %s)", oldpath, newpath)

	err := s.UnixFS.Link(oldpath, newpath)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

func (s *VerboseUnixFS) Stat(name string) (fs.FileInfo, error) {
	log.Infof("unixfs.Stat(%s)", name)

	fi, err := s.UnixFS.Stat(name)
	if err != nil {
		log.Warn(err)
		return nil, err
	}

	return fi, nil
}

func (s *VerboseUnixFS) Open(name string) (fs.File, error) {
	log.Infof("unixfs.Open(%s)", name)

	return s.OpenFile(name, os.O_RDONLY, os.FileMode(0o644))
}

func (s *VerboseUnixFS) OpenFile(name string, flag int, mode os.FileMode) (fs.File, error) {
	log.Infof("unixfs.OpenFile(%s, %d, %o)", name, flag, mode)

	f, err := s.UnixFS.OpenFile(name, flag, mode)
	if err != nil {
		log.Warn(err)
		return nil, err
	}

	return f, nil
}

func (s *VerboseUnixFS) GetFileId(fi fs.FileInfo) uint64 {
	log.Infof("unixfs.GetFileId: %#v", fi)

	inode := s.UnixFS.GetFileId(fi)
	log.Info("  id: ", inode)

	return inode
}

func (s *VerboseUnixFS) GetHandle(fi fs.FileInfo) ([]byte, error) {
	log.Infof("GetHandle: %s", fi.Name())

	b, err := s.UnixFS.GetHandle(fi)
	if err != nil {
		log.Warn(err)
		return nil, err
	}

	return b, nil
}

func (s *VerboseUnixFS) GetRootHandle() []byte {
	b := s.UnixFS.GetRootHandle()
	log.Infof("unixfs.GetRootHandle: %x", b)

	return b
}

// ResolveHandle resolves a file-handle(eg. nfs_fh4) to a full path name.
func (s *VerboseUnixFS) ResolveHandle(fh []byte) (string, error) {
	log.Infof("unixfs.ResolveHandle: %x", fh)

	v, err := s.UnixFS.ResolveHandle(fh)
	if err != nil {
		log.Warn(err)
		return "", err
	}

	return v, nil
}
