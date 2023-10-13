package implv4

import (
	"os"
	"testing"

	"github.com/smallfz/libnfs-go/nfs"
)

func TestAccess_file_0555_w(t *testing.T) {
	mode := os.FileMode(0o555)
	access := nfs.ACCESS4_MODIFY

	support, accForFh := computeAccessOnFile(mode, access)

	if (support & nfs.ACCESS4_MODIFY) == 0 {
		t.Fatalf("expects supporting writable. gets otherwise.")
	}

	if (accForFh & nfs.ACCESS4_MODIFY) > 0 {
		t.Fatalf("expects not writable to the file. gets otherwise.")
	}
}
