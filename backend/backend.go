package backend

import (
	"github.com/smallfz/libnfs-go/fs"
	"github.com/smallfz/libnfs-go/nfs"
)

type backendSession struct {
	vfs  fs.FS
	stat *Stat
}

func (s *backendSession) Close() error {
	if s.stat != nil {
		s.stat.CleanUp()
	}
	return nil
}

func (s *backendSession) GetFS() fs.FS {
	return s.vfs
}

func (s *backendSession) GetStatService() nfs.StatService {
	return s.stat
}

type Backend struct {
	vfs fs.FS
}

// New creates a new Backend instance.
func New(vfs fs.FS) *Backend {
	return &Backend{vfs: vfs}
}

func (b *Backend) CreateSession(state nfs.SessionState) nfs.BackendSession {
	return &backendSession{vfs: b.vfs, stat: new(Stat)}
}
