package memfs

import (
	"fmt"
	"github.com/smallfz/libnfs-go/fs"
	"github.com/smallfz/libnfs-go/log"
	"io"
)

type closeHandler func(bool, []byte)

type fileOpenFlags struct {
	trunc  bool
	append bool
}

func (f *fileOpenFlags) String() string {
	return fmt.Sprintf("{trunc=%v, append=%v}", f.trunc, f.append)
}

type memFile struct {
	s       *MemFS
	n       *memFsNode
	fi      *fileInfo
	buff    *Buffer
	changed bool
	onClose closeHandler
}

func newMemFile(s *MemFS, n *memFsNode, flag *fileOpenFlags, onClose closeHandler) *memFile {
	dat := []byte{}
	changed := false

	if !flag.trunc {
		dn := s.store.Get(n.nodeId)
		if dn != nil {
			if data, err := io.ReadAll(dn.Reader()); err == nil {
				dat = data
			}
		}
	} else {
		if n.size > 0 {
			changed = true
		}
		n.size = 0
	}

	buff := NewBuffer(dat)

	if flag.append {
		buff.Seek(0, io.SeekEnd)
	}

	return &memFile{
		s:       s,
		n:       n,
		fi:      s.getFileInfo(n),
		buff:    buff,
		changed: changed,
		onClose: onClose,
	}
}

func (f *memFile) Name() string {
	return f.fi.name
}

func (f *memFile) Stat() (fs.FileInfo, error) {
	return f.fi, nil
}

func (f *memFile) Read(buff []byte) (int, error) {
	if f.fi.IsDir() {
		return 0, io.EOF
	}
	log.Printf("memFile.Read(%d)", len(buff))
	i, err := f.buff.Read(buff)
	if err != nil {
		log.Warnf(" %v", err)
	} else {
		log.Printf(" -- %d bytes read.", i)
	}
	return i, err
}

func (f *memFile) Write(data []byte) (int, error) {
	if f.fi.IsDir() {
		return 0, io.EOF
	}
	log.Printf("memFile.Write(%d)", len(data))
	f.changed = true
	return f.buff.Write(data)
}

func (f *memFile) Seek(offset int64, whence int) (int64, error) {
	if f.fi.IsDir() {
		return 0, io.EOF
	}
	log.Printf("memFile.Seek(%d, %d)", offset, whence)
	return f.buff.Seek(offset, whence)
}

func (f *memFile) Truncate() error {
	log.Printf("memFile.Truncate()")
	f.buff.Truncate()
	f.n.size = int64(f.buff.size())
	log.Printf("  -- size after: %d", f.buff.size())
	return nil
}

func (f *memFile) Close() error {
	if f.onClose != nil {
		f.onClose(f.changed, f.buff.Bytes())
	}
	return nil
}

func (f *memFile) Sync() error {
	return nil
}

func (f *memFile) Readdir(n int) ([]fs.FileInfo, error) {
	if f.fi.IsDir() {
		fiList := []fs.FileInfo{}
		if f.n.children != nil {
			for _, child := range f.n.children {
				fiList = append(fiList, f.s.getFileInfo(child))
			}
		}
		return fiList, nil
	}
	return nil, io.EOF
}
