package xdr

import (
	"bytes"
	// "fmt"
	// "math"
	"encoding/binary"
	"testing"
	"time"
)

func TestWriter_varSizeBytes(t *testing.T) {
	buff := bytes.NewBuffer([]byte{})
	writer := NewWriter(buff)
	if size, err := writer.WriteAny([0]byte{}); err != nil {
		t.Fatalf("WriteAny(void): %v", err)
	} else if size != 0 {
		t.Fatalf("WriteAny(void): expects 0 bytes sent. but get %d.", size)
	}
}

func TestWriter_timeDuration(t *testing.T) {
	buff := bytes.NewBuffer([]byte{})
	writer := NewWriter(buff)

	dur := time.Hour * 4
	if size, err := writer.WriteAny(dur); err != nil {
		t.Fatalf("WriteAny(time.Duration): %v", err)
	} else if size != 8 {
		t.Fatalf("WriteAny(time.Duration): expects 8 bytes sent. but get %d.", size)
	}
}

func TestWriter_Struct(t *testing.T) {
	a := &structTestA{
		Iv:   123,
		Iv32: 456,
		Values: []*structTestB{
			{Value: 789},
			{Value: 10010},
		},
		SingleB: &structTestB{Value: 666},
		PlainB:  structTestB{Value: 999},
		Text:    "hello",
	}

	src := bytes.NewBuffer([]byte{})

	buff := make([]byte, 8)
	binary.BigEndian.PutUint64(buff, a.Iv)
	src.Write(buff)

	buff = make([]byte, 4)
	binary.BigEndian.PutUint32(buff, a.Iv32)
	src.Write(buff)

	buff = make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(len(a.Values)))
	src.Write(buff)

	for _, v := range a.Values {
		buff := make([]byte, 4)
		binary.BigEndian.PutUint32(buff, uint32(v.Value))
		src.Write(buff)
	}

	buff = make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(a.SingleB.Value))
	src.Write(buff)

	// .PlainB
	buff = make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(a.PlainB.Value))
	src.Write(buff)

	// .Text
	buff = make([]byte, 4)
	txtb := []byte(a.Text)
	binary.BigEndian.PutUint32(buff, uint32(len(txtb)))
	src.Write(buff)
	src.Write(txtb)
	padLen := Pad(len(txtb))
	pad := make([]byte, padLen)
	src.Write(pad)

	// --- ---
	target := bytes.NewBuffer([]byte{})
	writer := NewWriter(target)

	if size, err := writer.WriteAny(a); err != nil {
		t.Fatalf("WriteAny: %v", err)
	} else if size != len(src.Bytes()) {
		t.Fatalf("expectes size %d but get %d.", len(src.Bytes()), size)
	} else {
		if !bytes.Equal(src.Bytes(), target.Bytes()) {
			t.Fatalf(
				"expects: \n%v\nbut gets:\n%v\n",
				src.Bytes(), target.Bytes(),
			)
		}
	}
}
