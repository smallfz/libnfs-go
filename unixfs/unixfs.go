package unixfs

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"

	"github.com/smallfz/libnfs-go/fs"
)

type UnixFS struct {
	workdir    string
	inodes     *Inodes
	attributes fs.Attributes
}

func New(workdir string) (*UnixFS, error) {
	workdir, err := filepath.Abs(workdir)
	if err != nil {
		return nil, err
	}

	inodes := NewInodes()
	err = inodes.Scan(workdir)
	if err != nil {
		return nil, err
	}

	return &UnixFS{
		workdir: workdir,
		inodes:  inodes,
		attributes: fs.Attributes{
			LinkSupport:     true,
			SymlinkSupport:  true,
			ChownRestricted: true,  // Supported but chown is disabled by default for security reason
			MaxName:         255,   // Common value
			NoTrunc:         false, // Safe value
		},
	}, nil
}

func (s *UnixFS) Attributes() *fs.Attributes {
	return &s.attributes
}

func (s *UnixFS) Stat(name string) (fs.FileInfo, error) {
	name, err := s.ResolveUnix(name)
	if err != nil {
		return nil, err
	}

	return Stat(name)
}

func (s *UnixFS) Chmod(name string, mode os.FileMode) error {
	name, err := s.ResolveUnix(name)
	if err != nil {
		return err
	}

	return os.Chmod(name, mode)
}

func (s *UnixFS) Chown(name string, uid, gid int) error {
	name, err := s.ResolveUnix(name)
	if err != nil {
		return err
	}

	return os.Lchown(name, uid, gid)
}

func (s *UnixFS) MkdirAll(name string, mode os.FileMode) error {
	if name == fs.ROOT {
		return nil
	}

	name, err := s.ResolveUnix(name)
	if err != nil {
		return err
	}

	//

	if err = os.MkdirAll(name, mode); err != nil {
		return err
	}

	//

	return s.inodes.UpdateAll(name)
}

func (s *UnixFS) Remove(name string) error {
	if name == fs.ROOT {
		return os.ErrPermission
	}

	name, err := s.ResolveUnix(name)
	if err != nil {
		return err
	}

	//

	if err = os.Remove(name); err != nil {
		return err
	}

	//

	s.inodes.RemovePath(name)
	return nil
}

func (s *UnixFS) Rename(oldpath, newpath string) error {
	if oldpath == fs.ROOT || newpath == fs.ROOT || oldpath == newpath {
		return os.ErrPermission
	}

	oldpath, err := s.ResolveUnix(oldpath)
	if err != nil {
		return err
	}

	newpath, err = s.ResolveUnix(newpath)
	if err != nil {
		return err
	}

	//

	if err = os.Rename(oldpath, newpath); err != nil {
		return err
	}

	//

	s.inodes.RemovePath(oldpath)
	return s.inodes.UpdateAll(newpath)
}

func (s *UnixFS) Link(oldpath, newpath string) error {
	if oldpath == fs.ROOT || newpath == fs.ROOT || oldpath == newpath {
		return os.ErrPermission
	}

	oldpath, err := s.ResolveUnix(oldpath)
	if err != nil {
		return err
	}

	newpath, err = s.ResolveUnix(newpath)
	if err != nil {
		return err
	}

	//

	if err = os.Link(oldpath, newpath); err != nil {
		return err
	}

	//

	return s.inodes.UpdateAll(newpath)
}

func (s *UnixFS) Symlink(oldpath, newpath string) error {
	if oldpath == fs.ROOT || newpath == fs.ROOT || oldpath == newpath {
		return os.ErrPermission
	}

	newpath, err := s.ResolveUnix(newpath)
	if err != nil {
		return err
	}

	//

	// We do not check `oldpath', we let the OS put the data in the link file.
	if err = os.Symlink(oldpath, newpath); err != nil {
		return err
	}

	return s.inodes.UpdateAll(newpath)
}

func (s *UnixFS) Readlink(name string) (string, error) {
	name, err := s.ResolveUnix(name)
	if err != nil {
		return "", err
	}

	return os.Readlink(name)
}

func (s *UnixFS) Open(name string) (fs.File, error) {
	return s.OpenFile(name, os.O_RDONLY, os.FileMode(0o644))
}

func (s *UnixFS) OpenFile(name string, flag int, mode os.FileMode) (fs.File, error) {
	name, err := s.ResolveUnix(name)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(name, flag, mode)
	if err != nil {
		return nil, err
	}

	file := &file{
		File: *f,
	}

	{
		i, err := f.Stat()
		if err != nil {
			return nil, err
		}
		s.inodes.Add(statFromInfo(i).Inode(), name)
	}

	return file, nil
}

func (s *UnixFS) GetFileId(fi fs.FileInfo) uint64 {
	switch i := fi.(type) {
	case *FileInfo:
		return i.Inode()
	default:
		return InvalidID
	}
}

func (s *UnixFS) GetHandle(fi fs.FileInfo) ([]byte, error) {
	id := s.GetFileId(fi)
	if id == InvalidID {
		return nil, os.ErrNotExist
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, id)
	return buf, nil
}

func (s *UnixFS) GetRootHandle() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, s.inodes.GetID(s.workdir))
	return buf
}

// ResolveHandle resolves a file-handle(eg. nfs_fh4) to a full path name.
func (s *UnixFS) ResolveHandle(fh []byte) (string, error) {
	var id uint64
	if len(fh) <= 8 {
		id = binary.BigEndian.Uint64(fh)
	} else {
		id = binary.BigEndian.Uint64(fh[:8])
	}

	if id == InvalidID {
		return "", os.ErrNotExist
	}

	path := s.inodes.GetPath(id)
	if path == "" {
		return "", os.ErrNotExist
	}

	return s.ResolveNFS(path), nil
}

//
// Helpers
//

func (s *UnixFS) ResolveUnix(name string) (string, error) {
	name = fs.Abs(name)
	name = filepath.FromSlash(name)
	return filepath.Join(s.workdir, name), nil
}

func (s *UnixFS) ResolveNFS(name string) string {
	name = strings.TrimPrefix(name, s.workdir)
	name = filepath.ToSlash(name)
	if name == "" {
		name = "/"
	}

	return name
}
