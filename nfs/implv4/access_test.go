package implv4

import (
	"libnfs-go/nfs"
	"os"
	"testing"
)

func TestAccess_file_0555_w(t *testing.T) {
	mode := os.FileMode(0555)
	access := nfs.ACCESS4_MODIFY

	support, accForFh := computeAccessOnFile(mode, access)

	if (support & nfs.ACCESS4_MODIFY) == 0 {
		t.Fatalf("expects supporting writable. gets otherwise.")
	}

	if (accForFh & nfs.ACCESS4_MODIFY) > 0 {
		t.Fatalf("expects not writable to the file. gets otherwise.")
	}
}
