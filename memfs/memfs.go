// memfs includes an in-memory backend implementation for NFS server.
package memfs

import (
	"bytes"
	// "crypto/md5"
	// "encoding/binary"
	"io"
	"libnfs-go/fs"
	"libnfs-go/log"
	// "libnfs-go/log"
	"os"
	"path"
	"time"
	// "strings"
	"sync"
)

type memFsNode struct {
	id       int
	name     string
	isDir    bool
	nodeId   uint64 // storage data-node id
	perm     os.FileMode
	aTime    time.Time
	cTime    time.Time
	mTime    time.Time
	size     int64
	children []*memFsNode
}

func (n *memFsNode) findChild(parts []string) (*memFsNode, bool) {
	if n.children == nil || !n.isDir {
		return nil, false
	}
	for _, child := range n.children {
		if child.name == parts[0] {
			if len(parts) == 1 {
				return child, true
			}
			return child.findChild(parts[1:])
		}
	}
	return nil, false
}

func (n *memFsNode) addChild(child *memFsNode) error {
	if !n.isDir {
		return os.ErrPermission
	}

	if n.children == nil {
		n.children = []*memFsNode{}
	}

	n.children = append(n.children, child)
	n.mTime = time.Now()

	return nil
}

func (n *memFsNode) removeChild(name string) (*memFsNode, bool) {
	if !n.isDir || n.children == nil {
		return nil, false
	}

	target := (*memFsNode)(nil)
	for _, child := range n.children {
		if child.name == name {
			target = child
			break
		}
	}

	if target == nil {
		return nil, false
	}

	cList := []*memFsNode{}
	for _, child := range n.children {
		if child.name != target.name {
			cList = append(cList, child)
		}
	}

	n.children = cList
	n.mTime = time.Now()

	return target, true
}

// -----

type MemFS struct {
	store fs.Storage
	root  *memFsNode

	fileId int
	lck    *sync.RWMutex
}

func NewMemFS() *MemFS {
	store := NewStorage()
	return &MemFS{
		lck:   &sync.RWMutex{},
		store: store,
		root: &memFsNode{
			id:    1000,
			name:  "",
			isDir: true,
			perm:  os.FileMode(0755),
			cTime: time.Now(),
			mTime: time.Now(),
		},
	}
}

func (s *MemFS) nextId() int {
	s.lck.Lock()
	defer s.lck.Unlock()

	if s.fileId < 1000 {
		s.fileId = 1000
	}

	s.fileId++
	return s.fileId
}

func (s *MemFS) getFileInfo(n *memFsNode) *fileInfo {
	nlinks := 1
	if n.isDir {
		nlinks += 1
	}
	if n.isDir && n.children != nil {
		nlinks += len(n.children)
	}
	return &fileInfo{
		name:     path.Base(n.name),
		perm:     n.perm,
		size:     n.size,
		modTime:  n.mTime,
		aTime:    n.aTime,
		cTime:    n.cTime,
		isDir:    n.isDir,
		numLinks: nlinks,
	}
}

func (s *MemFS) getNode(name string) (*memFsNode, bool) {
	parts := fs.BreakAll(name)
	if len(parts) <= 0 {
		return s.root, true
	}
	return s.root.findChild(parts)
}

func (s *MemFS) findNodes(prefix string) []*memFsNode {
	rs := []*memFsNode{}
	n, found := s.getNode(prefix)
	if !found {
		return rs
	}
	if n.isDir && n.children != nil {
		rs = append(rs, n.children...)
	}
	return rs
}

func (s *MemFS) writeNode(n *memFsNode, dat []byte) {
	log.Warnf("%T.writeNode(%s, %d bytes)", s, n.name, len(dat))
	src := bytes.NewReader(dat)
	src.Seek(0, io.SeekStart)
	dn := s.store.Get(n.nodeId)
	if dn == nil {
		nodeId, err := s.store.Create(src)
		if err != nil {
			return
		}
		n.nodeId = nodeId
	} else {
		s.store.Update(n.nodeId, src)
	}
	n.mTime = time.Now()
	n.cTime = time.Now()
	n.size = int64(s.store.Size(n.nodeId))
}

func (s *MemFS) Open(name string) (fs.File, error) {
	return s.OpenFile(name, os.O_RDONLY, os.FileMode(0644))
}

func (s *MemFS) OpenFile(name string, flag int, perm os.FileMode) (fs.File, error) {
	parts := fs.BreakAll(name)

	autoCreate := (flag & os.O_CREATE) > 0
	overwrite := (flag & os.O_EXCL) == 0

	n, found := s.getNode(name)
	if !found {
		if !autoCreate {
			return nil, os.ErrNotExist
		}

		folderPath := fs.Abs(fs.Join(parts[:len(parts)-1]...))
		folder, found := s.getNode(folderPath)
		if !found {
			return nil, os.ErrPermission
		}

		baseName := parts[len(parts)-1]

		n := &memFsNode{
			id:    s.nextId(),
			name:  baseName,
			isDir: false,
			perm:  perm,
			cTime: time.Now(),
			mTime: time.Now(),
		}
		if err := folder.addChild(n); err != nil {
			return nil, err
		}

		log.Printf("MemFS.OpenFile: new file(name=%s) created.", baseName)

		flags := &fileOpenFlags{
			trunc:  false,
			append: (flag & os.O_APPEND) > 0,
		}
		return newMemFile(s, n, flags, func(changed bool, dat []byte) {
			n.aTime = time.Now()
			s.writeNode(n, dat)
		}), nil

	}

	flagW := os.O_WRONLY | os.O_RDWR | os.O_APPEND | os.O_TRUNC | os.O_CREATE
	writing := (flag & flagW) > 0
	if writing && !overwrite {
		return nil, os.ErrExist
	}

	log.Printf("MemFS.OpenFile: flag = %v", flag)
	log.Printf("  O_RDONLY: %v", os.O_RDONLY&flag)
	log.Printf("  O_WRONLY: %v", os.O_WRONLY&flag)
	log.Printf("  O_RDWR: %v", os.O_RDWR&flag)
	log.Printf("  O_APPEND: %v", os.O_APPEND&flag)
	log.Printf("  O_CREATE: %v", os.O_CREATE&flag)
	log.Printf("  O_EXCL: %v", os.O_EXCL&flag)
	log.Printf("  O_SYNC: %v", os.O_SYNC&flag)
	log.Printf("  O_TRUNC: %v", os.O_TRUNC&flag)

	trunc := (flag & os.O_TRUNC) > 0

	flags := &fileOpenFlags{
		trunc:  trunc,
		append: (flag & os.O_APPEND) > 0,
	}

	return newMemFile(s, n, flags, func(changed bool, dat []byte) {
		n.aTime = time.Now()
		if !writing || !changed {
			// log.Printf("MemFS.OpenFile: not with writing modes. discard writing.")
			return
		}
		// log.Printf("MemFS.OpenFile: writing %d bytes.", len(dat))
		s.writeNode(n, dat)
	}), nil

}

func (s *MemFS) Stat(name string) (fs.FileInfo, error) {
	n, found := s.getNode(name)
	if !found {
		return nil, os.ErrNotExist
	}
	return s.getFileInfo(n), nil
}

func (s *MemFS) Chmod(name string, perm os.FileMode) error {
	n, found := s.getNode(name)
	if !found {
		return os.ErrNotExist
	}

	mask := (uint32(1) << 24) - 1
	p := uint32(perm) & mask

	typ := ((uint32(1) << 8) - 1) << 24
	typ = typ & uint32(n.perm)

	n.perm = os.FileMode(typ | p)

	n.mTime = time.Now()
	n.cTime = time.Now()
	return nil
}

func (s *MemFS) Rename(oldName, newName string) error {
	name := fs.Abs(oldName)
	newName = fs.Abs(newName)

	if name == fs.ROOT || newName == fs.ROOT || name == newName {
		return os.ErrPermission
	}

	folder, _ := path.Split(name)
	folderDst, _ := path.Split(newName)

	if folder != folderDst {
		return os.ErrPermission
	}

	if _, found := s.getNode(newName); found {
		return os.ErrExist
	}

	n, found := s.getNode(name)
	if !found {
		return os.ErrNotExist
	} else {
		n.name = fs.Base(newName)
	}

	return nil
}

func (s *MemFS) Remove(name string) error {
	name = fs.Abs(name)

	if name == fs.ROOT {
		return os.ErrPermission
	}

	nHit, found := s.getNode(name)
	if !found {
		return os.ErrNotExist
	}

	if nHit.isDir {
		if nHit.children != nil && len(nHit.children) > 0 {
			// has children, no cascading removing allowed
			return os.ErrExist
		}
	}

	folderPath := fs.Dir(name)
	folder, found := s.getNode(folderPath)
	if !found {
		return os.ErrPermission
	}

	folder.removeChild(fs.Base(name))

	return nil
}

func (s *MemFS) MkdirAll(name string, perm os.FileMode) error {
	name = fs.Abs(name)

	if name == fs.ROOT {
		return nil
	}

	n, found := s.getNode(name)
	if found {
		return os.ErrExist
	}

	folderPath := fs.Dir(name)
	folder, found := s.getNode(folderPath)
	if !found {
		if err := s.MkdirAll(folderPath, perm); err != nil {
			if !os.IsExist(err) {
				return err
			}
		}
		folder, found = s.getNode(folderPath)
		if !found {
			// looks like failed creating parent.
			return os.ErrPermission
		}
	}

	n = &memFsNode{
		id:    s.nextId(),
		name:  fs.Base(name),
		isDir: true,
		perm:  perm,
		cTime: time.Now(),
		mTime: time.Now(),
	}
	folder.addChild(n)

	return nil
}

func (s *MemFS) GetHandle(name string) ([]byte, error) {
	dat := []byte(fs.Abs(name))
	return dat, nil
}

func (s *MemFS) GetFileId(name string) uint64 {
	name = fs.Abs(name)
	n, found := s.getNode(name)
	if !found {
		return 0
	}
	return uint64(n.id)
}
