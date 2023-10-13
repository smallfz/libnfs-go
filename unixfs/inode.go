package unixfs

import (
	"io/fs"
	"path/filepath"
	"sync"

	"github.com/smallfz/libnfs-go/log"
	"github.com/smallfz/libnfs-go/memfs"
)

const InvalidID = memfs.InvalidId

// Inodes contain the mapping of paths and their inode IDs because
// it's more safe to rely on them instead of other kind of IDs.
type Inodes struct {
	mu     sync.Mutex
	inodes map[uint64]string
	paths  map[string]uint64
}

func NewInodes() *Inodes {
	return &Inodes{
		inodes: make(map[uint64]string),
		paths:  make(map[string]uint64),
	}
}

// Scan reads all the inode's IDs under the given workdir.
// It can take a while depending on the number of inodes to read and disk perfomance.
func (i *Inodes) Scan(workdir string) error {
	// On macOS we can avoid scanning all the workdir by using these 2 commands:
	//
	// [~]>> stat $HOME
	// 16777220 30109 drwxr-xr-x 75 mdouchement staff 0 2400 "Sep 14 10:07:41 2023" "Sep 14 10:07:51 2023" "Sep 14 10:07:51 2023" "Aug 10 15:49:15 2022" 4096 0 0 /Users/mdouchement
	// [~]>> GetFileInfo /.vol/16777220/30109
	// directory: "/Users/mdouchement"
	// attributes: avbstclinmedz
	// created: 08/10/2022 15:49:15
	// modified: 09/14/2023 10:08:02
	//
	// For the moment, no equivalent feature has been found on GNU/Linux.

	i.mu.Lock()
	defer i.mu.Unlock()

	log.Info("Scanning inodes...")

	return filepath.WalkDir(workdir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		inode := statFromInfo(info).Inode()
		i.inodes[inode] = path
		i.paths[path] = inode
		log.Debug("  ", inode, " ", path)
		return nil
	})
}

func (i *Inodes) UpdateAll(name string) error {
	var notfirst bool // Refresh full path because it may change due to UnixFS actions.

	for {
		if notfirst && i.ExistPath(name) {
			// The workdir (root folder) exists so we should never go outside the workdir.
			break
		}

		fi, err := Stat(name)
		if err != nil {
			return err
		}

		i.Add(fi.Inode(), name)

		name = filepath.Dir(name)
		notfirst = true
	}

	return nil
}

func (i *Inodes) GetPath(inode uint64) string {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.inodes[inode]
}

func (i *Inodes) GetID(path string) uint64 {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.paths[path]
}

func (i *Inodes) ExistPath(path string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	_, ok := i.paths[path]
	return ok
}

func (i *Inodes) Add(inode uint64, path string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.inodes[inode] = path
	i.paths[path] = inode
}

func (i *Inodes) RemoveID(inode uint64) {
	i.mu.Lock()
	defer i.mu.Unlock()

	path := i.inodes[inode]
	delete(i.inodes, inode)
	delete(i.paths, path)
}

func (i *Inodes) RemovePath(path string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	inode := i.paths[path]
	delete(i.inodes, inode)
	delete(i.paths, path)
}
