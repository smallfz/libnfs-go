package memfs

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"
)

func TestBufferReadingCopyN(t *testing.T) {
	src := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	buff := NewBuffer(src)

	input := bytes.NewBuffer([]byte{200, 201, 202})

	if size, err := io.CopyN(buff, input, 3); err != nil {
		t.Fatalf("io.CopyN: %v", err)
	} else if size != 3 {
		t.Fatalf("unexpected copied size %d.", size)
	}
}

func TestBufferReading(t *testing.T) {
	src := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	buff := NewBuffer(src)

	if _, err := buff.Seek(2, io.SeekStart); err != nil {
		t.Fatalf("buff.Seek: %v", err)
		return
	}

	dat := make([]byte, 3)

	if size, err := buff.Read(dat); err != nil {
		t.Fatalf("buff.Read: %v", err)
	} else if size != len(dat) {
		t.Fatalf(
			"unexpected size %d returned from Read. expects %d.",
			size, len(dat),
		)
	} else if !bytes.Equal(dat, src[2:5]) {
		t.Fatalf("unexpected reading result: %v", dat)
	}
}

func TestBufferSeekingHole(t *testing.T) {
	buff := NewBuffer([]byte{})
	buff.Write([]byte("0123456789"))

	buff.Seek(4, io.SeekStart)
	buff.Write([]byte("____"))

	rs := string(buff.Bytes())
	want := "0123____89"
	if rs != want {
		t.Fatalf("expects '%s' but gets '%s'.", want, rs)
		return
	}

	buff.Seek(20, io.SeekStart)
	buff.Write([]byte("AaAa"))

	fmt.Printf("%v", buff.Bytes())

	buff.Seek(10, io.SeekStart)
	buff.Write([]byte("-+-+-+****"))

	rs = string(buff.Bytes())
	want = "0123____89-+-+-+****AaAa"
	if rs != want {
		t.Fatalf("expects '%s' but gets '%s'.", want, rs)
		return
	}
}

type writingBlock struct {
	offset int64
	data   []byte
}

func TestBufferSeekingMultipleTimes(t *testing.T) {
	blocks := []*writingBlock{
		&writingBlock{offset: 0, data: []byte("0123")},
		&writingBlock{offset: 4, data: []byte("4567")},
		&writingBlock{offset: 8, data: []byte("89ab")},
		&writingBlock{offset: 12, data: []byte("cdef")},
		&writingBlock{offset: 16, data: []byte("____")},
		&writingBlock{offset: 20, data: []byte("AAA")},
	}

	index := make([]int, len(blocks))
	for i, _ := range index {
		index[i] = i
	}

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 20; i++ {
		buff := NewBuffer([]byte{})

		rand.Shuffle(len(index), func(x, y int) {
			index[x], index[y] = index[y], index[x]
		})

		fmt.Println(index)

		for _, v := range index {
			block := blocks[v]

			buff.Seek(block.offset, io.SeekStart)
			buff.Write(block.data)
		}

		rs := string(buff.Bytes())
		want := "0123456789abcdef____AAA"
		if rs != want {
			t.Fatalf("expects '%s' but gets '%s'.", want, rs)
			return
		}
	}
}
