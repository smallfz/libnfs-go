package implv4

import (
	"encoding/json"
	"testing"

	"github.com/smallfz/libnfs-go/log"
)

func TestFAttr4Decoding(t *testing.T) {
	a0 := newFAttr4(
		[]uint32{1575194, 11575866},
		"AAAAAQAAAABiCghRAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIgvZEdjdmJtOXliV0ZzTDNOeWN3PT0vc3JzL2RhdGEvc3IvNmQwOTNjNzgzZTg2MTFlYmI2MWVjYWU4YTgyOTA4MTUvU2NyZWVuc2hvdF8yMDIwMTIxNV8xMTMwMDJfY29tLmNoaW5hdW5pY29tLmdlYXNzaXN0YW50XzE2MDgwMDMyNjQuanBnAAAAAAAAAaYAAAGkAAAAAQAAAAEwAAAAAAAAATAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAGIKCFEAAAAAAAAAAGIKCFEAAAAAAAAAAGIKCFEAAAAAAAAAAAAAAaY=",
	)

	a1 := newFAttr4(
		[]uint32{1575194, 11575866},
		"AAAAAQAAAABiCimLAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAHMvc3JzL2RhdGEvc3IvNmQwOTNjNzgzZTg2MTFlYmI2MWVjYWU4YTgyOTA4MTUvU2NyZWVuc2hvdF8yMDIwMTIxNV8xMTMwMDJfY29tLmNoaW5hdW5pY29tLmdlYXNzaXN0YW50XzE2MDgwMDMyNjQuanBnAAAAAAAAAAT0AAAB7QAAAAEAAAABMAAAAAAAAAEwAAAAAAAAAAAAAAAAAAAAAAAQAAAAAABiCimLAAAAAAAAAABiCimLAAAAAAAAAABiCimLAAAAAAAAAAAAAAT0",
	)

	if aa0, err := decodeFAttrs4(a0); err != nil {
		t.Fatalf("decodeFAttrs4(a0): %v", err)
	} else {
		if dat, err := json.MarshalIndent(aa0, "", "  "); err != nil {
			t.Fatalf("json.MarshalIndent: %v", err)
		} else {
			log.Println("aa0:")
			log.Println(string(dat))
		}
	}

	if aa1, err := decodeFAttrs4(a1); err != nil {
		t.Fatalf("decodeFAttrs4(a1): %v", err)
	} else {
		if dat, err := json.MarshalIndent(aa1, "", "  "); err != nil {
			t.Fatalf("json.MarshalIndent: %v", err)
		} else {
			log.Println("aa1:")
			log.Println(string(dat))
		}
	}
}
