package memfs

import (
	"libnfs-go/fs"
	"libnfs-go/nfs"
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
	if s.stat == nil {
		s.stat = &Stat{}
	}
	return s.stat
}

type Backend struct {
	vfs fs.FS
}

// NewBackend creates a new Backend instalce.
func NewBackend(vfs fs.FS) *Backend {
	return &Backend{vfs: vfs}
}

func (b *Backend) CreateSession(state nfs.SessionState) nfs.BackendSession {
	return &backendSession{vfs: b.vfs}
}
