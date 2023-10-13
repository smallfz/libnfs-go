package xdr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
)

func TestReaderRead(t *testing.T) {
	srcb := []byte{1, 2, 3, 4, 5, 6}
	src := bytes.NewBuffer(srcb)

	reader := NewReader(src)
	dst := make([]byte, 4)
	if size, err := reader.Read(dst); err != nil {
		t.Fatalf("Read: %v", err)
	} else if size != len(dst) {
		t.Fatalf("expects 4 bytes. get %d.", size)
	}
	if !bytes.Equal(dst, srcb[:4]) {
		t.Fatalf("expects %v but get %v.", srcb[:4], dst)
	}
}

func TestReaderReadBytes(t *testing.T) {
	srcb := []byte{1, 2, 3, 4, 5, 6}
	src := bytes.NewBuffer(srcb)

	reader := NewReader(src)
	size := 4
	if dst, err := reader.ReadBytes(size); err != nil {
		t.Fatalf("ReadBytes: %v", err)
	} else if size != len(dst) {
		t.Fatalf("expects 4 bytes. get %d.", size)
	} else {
		if !bytes.Equal(dst, srcb[:4]) {
			t.Fatalf("expects %v but get %v.", srcb[:4], dst)
		}
	}
}

func TestReaderReadAs_bool(t *testing.T) {
	srcb := []byte{0, 0, 0, 1}
	src := bytes.NewBuffer(srcb)

	reader := NewReader(src)

	target := false
	if size, err := reader.ReadAs(&target); err != nil {
		t.Fatalf("ReadAs: %v", err)
	} else if size != 4 {
		t.Fatalf("expects 4 bytes. get %d.", size)
	} else {
		if !target {
			t.Fatalf("expects TRUE but get %v.", target)
		}
	}

	copy(srcb, []byte{0, 0, 0, 0, 0, 0})
	src = bytes.NewBuffer(srcb)
	reader = NewReader(src)

	if size, err := reader.ReadAs(&target); err != nil {
		t.Fatalf("ReadAs: %v", err)
	} else if size != 4 {
		t.Fatalf("expects 4 bytes. get %d.", size)
	} else {
		if target {
			t.Fatalf("expects FALSE but get %v.", target)
		}
	}
}

func TestReaderReadAs_uint16(t *testing.T) {
	srcb := []byte{0, 0, 1, 0}
	src := bytes.NewBuffer(srcb)
	srcVal := binary.BigEndian.Uint32(srcb)

	reader := NewReader(src)

	target := uint16(0)
	if size, err := reader.ReadAs(&target); err != nil {
		t.Fatalf("ReadAs: %v", err)
	} else if size != 4 {
		t.Fatalf("expects 4 bytes. get %d.", size)
	} else {
		if uint32(target) != srcVal {
			t.Fatalf("expects %d but get %d.", srcVal, target)
		}
	}
}

func TestReaderReadAs_uint64(t *testing.T) {
	srcb := make([]byte, 32)
	srcVal := uint64(3721352)
	binary.BigEndian.PutUint64(srcb, srcVal)

	src := bytes.NewBuffer(srcb)

	reader := NewReader(src)

	target := uint64(0)
	if size, err := reader.ReadAs(&target); err != nil {
		t.Fatalf("ReadAs: %v", err)
	} else if size != 8 {
		t.Fatalf("expects 8 bytes. get %d.", size)
	} else {
		if target != srcVal {
			t.Fatalf("expects %d but get %d.", srcVal, target)
		}
	}
}

func TestReaderReadAs_float32(t *testing.T) {
	srcb := []byte{0, 0, 0, 0}
	srcVal := float32(12.345)
	binary.BigEndian.PutUint32(srcb, math.Float32bits(srcVal))
	src := bytes.NewBuffer(srcb)

	reader := NewReader(src)

	target := float32(0)
	if size, err := reader.ReadAs(&target); err != nil {
		t.Fatalf("ReadAs: %v", err)
	} else if size != 4 {
		t.Fatalf("expects 4 bytes. get %d.", size)
	} else {
		if target != srcVal {
			t.Fatalf("expects %f but get %f.", srcVal, target)
		}
	}
}

func TestReaderReadAs_float64(t *testing.T) {
	srcb := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	srcVal := float64(12.345)
	binary.BigEndian.PutUint64(srcb, math.Float64bits(srcVal))
	src := bytes.NewBuffer(srcb)

	reader := NewReader(src)

	target := float64(0)
	if size, err := reader.ReadAs(&target); err != nil {
		t.Fatalf("ReadAs: %v", err)
	} else if size != 8 {
		t.Fatalf("expects 8 bytes. get %d.", size)
	} else {
		if target != srcVal {
			t.Fatalf("expects %f but get %f.", srcVal, target)
		}
	}
}

func TestReaderReadAs_varSizeArr(t *testing.T) {
	srcVal := []uint64{123, 456, 8, 9, 0, 999}
	src := bytes.NewBuffer([]byte{})

	// 1st word is array size
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(len(srcVal)))
	src.Write(buff)

	for _, v := range srcVal {
		buff := make([]byte, 8)
		binary.BigEndian.PutUint64(buff, v)
		src.Write(buff)
	}

	reader := NewReader(bytes.NewReader(src.Bytes()))

	target := []uint64{}
	if size, err := reader.ReadAs(&target); err != nil {
		t.Fatalf("ReadAs: %v", err)
	} else if size != 4+8*len(srcVal) {
		t.Fatalf("expects %d bytes. get %d.", 4+8*len(srcVal), size)
	} else {
		if len(target) != len(srcVal) {
			t.Fatalf("expects %d elements but get %d.", len(srcVal), len(target))
		} else {
			for i := 0; i < len(srcVal); i++ {
				if srcVal[i] != target[i] {
					t.Fatalf(
						"element[%d]: expects %d but get %d.",
						i, srcVal[i], target[i],
					)
				}
			}
		}
	}
}

func TestReaderReadAs_fixedSizeArr(t *testing.T) {
	srcVal := [6]uint64{123, 456, 8, 9, 0, 999}
	src := bytes.NewBuffer([]byte{})

	for _, v := range srcVal {
		buff := make([]byte, 8)
		binary.BigEndian.PutUint64(buff, v)
		src.Write(buff)
	}

	reader := NewReader(bytes.NewReader(src.Bytes()))

	target := [6]uint64{}
	if size, err := reader.ReadAs(&target); err != nil {
		t.Fatalf("ReadAs: %v", err)
	} else if size != 8*len(srcVal) {
		t.Fatalf("expects %d bytes. get %d.", 8*len(srcVal), size)
	} else {
		if len(target) != len(srcVal) {
			t.Fatalf("expects %d elements but get %d.", len(srcVal), len(target))
		} else {
			for i := 0; i < len(srcVal); i++ {
				if srcVal[i] != target[i] {
					t.Fatalf(
						"element[%d]: expects %d but get %d.",
						i, srcVal[i], target[i],
					)
				}
			}
		}
	}
}

func TestReaderReadAs_varSizeBytes(t *testing.T) {
	src := bytes.NewBuffer([]byte{})
	srcb := []byte("hello world")

	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(len(srcb)))
	src.Write(buff)

	src.Write(srcb)
	size := len(srcb)
	if size%4 != 0 {
		padLen := Pad(size)
		pad := make([]byte, padLen)
		src.Write(pad)
	}

	fmt.Println(src.Bytes())

	reader := NewReader(bytes.NewReader(src.Bytes()))

	target := ([]byte)(nil) // {}
	if size, err := reader.ReadAs(&target); err != nil {
		t.Fatalf("ReadAs: %v", err)
	} else if size != len(src.Bytes()) {
		t.Fatalf("expects %d bytes. get %d.", len(src.Bytes()), size)
	} else {
		if len(target) < len(srcb) {
			t.Fatalf("expects %d elements but get %d.", len(srcb), len(target))
		} else {
			target0 := target[0:len(srcb)]
			if !bytes.Equal(target0, srcb) {
				t.Fatalf("expects %v but get %v.", srcb, target0)
			}
			fmt.Println(target0)
		}
	}
}

func TestReaderReadAs_fizedSizeBytes(t *testing.T) {
	src := bytes.NewBuffer([]byte{})
	srcb := []byte("hello world")

	src.Write(srcb)
	size := len(srcb)
	if size%4 != 0 {
		padLen := Pad(size)
		pad := make([]byte, padLen)
		src.Write(pad)
	}

	fmt.Println(src.Bytes())

	reader := NewReader(bytes.NewReader(src.Bytes()))

	target := [11]byte{}
	if size, err := reader.ReadAs(&target); err != nil {
		t.Fatalf("ReadAs: %v", err)
	} else if size != len(src.Bytes()) {
		t.Fatalf("expects %d bytes. get %d.", len(src.Bytes()), size)
	} else {
		if len(target) < len(srcb) {
			t.Fatalf("expects %d elements but get %d.", len(srcb), len(target))
		} else {
			target0 := target[0:len(srcb)]
			if !bytes.Equal(target0, srcb) {
				t.Fatalf("expects %v but get %v.", srcb, target0)
			}
			fmt.Println(target0)
		}
	}
}

type structTestB struct {
	Value int
}

type structTestA struct {
	Iv      uint64
	Iv32    uint32
	Values  []*structTestB
	SingleB *structTestB
	PlainB  structTestB
	Text    string
}

func TestReaderReadAs_struct(t *testing.T) {
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

	// --- writing ends. ---

	srcb := src.Bytes()
	srcSize := len(srcb)

	reader := NewReader(bytes.NewReader(srcb))

	target := &structTestA{}
	if size, err := reader.ReadAs(target); err != nil {
		t.Fatalf("ReadAs: %v", err)
	} else if size != srcSize {
		t.Fatalf("expects %d bytes. get %d.", srcSize, size)
	} else {
		if target.Iv != a.Iv {
			t.Fatalf("target.Iv: expects %d but get %d.", a.Iv, target.Iv)
		}
		if target.Iv32 != a.Iv32 {
			t.Fatalf("target.Iv32: expects %d but get %d.", a.Iv32, target.Iv32)
		}
		if target.Values == nil || len(target.Values) != len(a.Values) {
			t.Fatalf("target.Values: expects %d elements. get %d.",
				len(a.Values), len(target.Values),
			)
		}
		if target.SingleB == nil || target.SingleB.Value != a.SingleB.Value {
			t.Fatalf(
				"target.SingleB.Value: expects %v but get %v.",
				a.SingleB, target.SingleB,
			)
		}
		if target.PlainB.Value != a.PlainB.Value {
			t.Fatalf(
				"target.PlainB.Value: expects %v but get %v.",
				a.PlainB, target.PlainB,
			)
		}
		if target.Text != a.Text {
			t.Fatalf(
				"target.Text: expects %s but get %s.",
				a.Text, target.Text,
			)
		}
	}
}
