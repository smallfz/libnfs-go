package implv4

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/smallfz/libnfs-go/nfs"
	"github.com/smallfz/libnfs-go/xdr"
)

func TestParsingSETCLIENTID4_res(t *testing.T) {
	raw, _ := base64.StdEncoding.DecodeString("J4bWCgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAIwAAAAADsSBgqrPnOuG/lGHO4oWS")

	reader := xdr.NewReader(bytes.NewBuffer(raw))

	_, err := reader.ReadUint32() /* xid */
	if err != nil {
		t.Fatalf("%v", err)
		return
	}

	msgType, err := reader.ReadUint32()
	if err != nil {
		t.Fatalf("%v", err)
		return
	}

	if msgType != xdr.RPC_REPLY {
		t.Fatalf("expects RPC_REPLY but get %d", msgType)
		return
	}

	replyStat, err := reader.ReadUint32()
	if err != nil {
		t.Fatalf("%v", err)
		return
	}

	if replyStat != nfs.ACCEPT_SUCCESS {
		t.Fatalf(" reply-stat: %d", replyStat)
		return
	}

	auth := &nfs.Auth{}
	if _, err := reader.ReadAs(auth); err != nil {
		t.Fatalf("%v", err)
		return
	}

	acceptStatus, err := reader.ReadUint32()
	if err != nil {
		t.Fatalf("%v", err)
		return
	}

	if acceptStatus != nfs.ACCEPT_SUCCESS {
		t.Fatalf(" accept-status: %d", acceptStatus)
		return
	}

	status, err := reader.ReadUint32()
	if err != nil {
		t.Fatalf(" reader.ReadUint32() => status: %v", err)
		return
	}

	if status != nfs.NFS4_OK {
		t.Fatalf(" status: %d", status)
	}

	res := &nfs.SETCLIENTID4resok{}
	if _, err := reader.ReadAs(res); err != nil {
		t.Fatalf("%v", err)
	}

	fmt.Println(toJson(res))
}
