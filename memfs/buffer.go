package memfs

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Buffer struct {
	data   []byte
	cur    int
	closed bool
	lck    *sync.RWMutex
}

func NewBuffer(src []byte) *Buffer {
	return &Buffer{data: src}
}

func (b *Buffer) init() {
	if b.lck == nil {
		b.lck = &sync.RWMutex{}
	}

	b.lck.Lock()
	defer b.lck.Unlock()

	if b.data == nil {
		b.data = []byte{}
		b.cur = 0
	}
}

func (b *Buffer) Size() int64 {
	b.init()
	b.lck.RLock()
	defer b.lck.RUnlock()
	return int64(b.size())
}

func (b *Buffer) size() int {
	if b.data != nil {
		return len(b.data)
	}
	return 0
}

func (b *Buffer) Bytes() []byte {
	b.init()
	b.lck.RLock()
	defer b.lck.RUnlock()
	return b.data[:]
}

func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	b.init()
	b.lck.Lock()
	defer b.lck.Unlock()

	if b.closed {
		return 0, io.EOF
	}

	size := b.size()
	cur := b.cur

	switch whence {
	case io.SeekStart:
		if offset == 0 && b.cur == 0 {
			return 0, nil
		}
		cur = int(offset)
		break

	case io.SeekCurrent:
		cur = b.cur + int(offset)
		break

	case io.SeekEnd:
		cur = size + int(offset)
		break
	}

	if cur < 0 {
		return int64(b.cur), fmt.Errorf("invalid seeking position.")
	}

	b.cur = cur
	return int64(b.cur), nil
}

func (b *Buffer) Write(dat []byte) (int, error) {
	if dat == nil || len(dat) == 0 {
		return 0, nil
	}

	b.init()
	b.lck.Lock()
	defer b.lck.Unlock()

	if b.closed {
		return 0, io.EOF
	}

	// extend
	extend := b.cur + len(dat) - b.size()
	if extend > 0 {
		b.data = append(b.data, make([]byte, extend)...)
	}

	// write
	count := copy(b.data[b.cur:], dat)
	b.cur += count

	return count, nil
}

func (b *Buffer) Read(dat []byte) (int, error) {
	if dat == nil || len(dat) == 0 {
		return 0, nil
	}

	b.init()
	b.lck.Lock()
	defer b.lck.Unlock()

	if b.closed {
		return 0, io.EOF
	}

	size := b.size()
	if b.cur >= size {
		return 0, io.EOF
	}

	end := b.cur + len(dat)
	if end > b.size() {
		end = b.size()
	}

	count := copy(dat, b.data[b.cur:end])
	b.cur = end

	return count, nil
}

func (b *Buffer) Truncate() {
	b.init()
	b.lck.Lock()
	defer b.lck.Unlock()

	if b.closed {
		return
	}

	size := b.size()
	if size <= 0 {
		return
	}

	if b.data != nil {
		b.data = b.data[0:b.cur]
	}
}

func (b *Buffer) Close() error {
	b.init()
	b.lck.Lock()
	defer b.lck.Unlock()

	if b.closed {
		return os.ErrClosed
	}

	if len(b.data) > 0 {
		b.data = b.data[0:0]
	}
	b.cur = 0

	b.closed = true

	return nil
}
