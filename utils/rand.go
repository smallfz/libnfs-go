package utils

import (
	"crypto/rand"
	"encoding/binary"
)

func RandUint32() uint32 {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return binary.BigEndian.Uint32(b)
}
