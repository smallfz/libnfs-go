package implv4

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"libnfs-go/nfs"
	"libnfs-go/xdr"
	"testing"
)

/*

C: setclientid
S: jbKrXwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAIwAAAAADsSBgc7fnOp8ZlWGa5oWS

C: setclientid_confirm
S: jrKrXwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAJAAAAAA=

C: putfh(fh="AQABAAAAAAA=") + access(Access=31) + getattr(args=[24, 3145728])
S: K3qbaAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAFgAAAAAAAAADAAAAAAAAAB8AAAAfAAAACQAAAAAAAAACAAAAGAAwAAAAAAAoYCC1bDi7FDcAAAAAAAAQAAAAAABgILVsOLsUNwAAAABgILVsOLsUNw==

C: putfh(fh="AQABAAAAAAA=") + readdir(args={
  "Cookie": 0,
  "CookieVerf": 0,
  "DirCount": 8170,
  "MaxCount": 32680,
  "AttrRequest": [
    1575194,
    11575866
  ]
})

S: LHqbaAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAFgAAAAAAAAAaAAAAAAAAAAAAAAAAAAAAAW+ZZ6tVAgIjAAAAA29yZwAAAAACABgJGgCwojoAAACYAAAAAmAgtWw7LK/JAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABABAAEBAAAAAAQANAAYTH4XAAAAAAA0AAQAAAHAAAAAAwAAAAEwAAAAAAAAATAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAGFyMN4CqqZ6AAAAAGAgtWw7LK/JAAAAAGAgtWw7LK/JAAAAAAA0AAQAAAABf/////////8AAAAHb3JnLXN2YwAAAAACABgJGgCwojoAAACYAAAAAmAgs54P1yYeAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABABAAEBAAAAAAIANADNxd4fAAAAAAA0AAIAAAHAAAAAAwAAAAEwAAAAAAAAATAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAGFyMN4CqqZ6AAAAAGAgs54P1yYeAAAAAGAgs54P1yYeAAAAAAA0AAIAAAAAAAAAAQ==

*/

func TestParsingCOMPOUND4_res_putfh_readdir(t *testing.T) {
	raw, _ := base64.StdEncoding.DecodeString("uNo+UAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAFgAAAAAAAAAaAAAAAAAAAAAAAAAAAAAAAW+ZZ6tVAgIjAAAAA29yZwAAAAACABgJGgCwojoAAACYAAAAAmAgtWw7LK/JAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABABAAEBAAAAAAQANAAYTH4XAAAAAAA0AAQAAAHAAAAAAwAAAAEwAAAAAAAAATAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAGGXaXYdnd+oAAAAAGAgtWw7LK/JAAAAAGAgtWw7LK/JAAAAAAA0AAQAAAABf/////////8AAAAHb3JnLXN2YwAAAAACABgJGgCwojoAAACYAAAAAmAgs54P1yYeAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABABAAEBAAAAAAIANADNxd4fAAAAAAA0AAIAAAHAAAAAAwAAAAEwAAAAAAAAATAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAGGXbKkygFszAAAAAGAgs54P1yYeAAAAAGAgs54P1yYeAAAAAAA0AAIAAAAAAAAAAQ==")

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

	// decode compound result...

	tag := ""
	if _, err := reader.ReadAs(&tag); err != nil {
		t.Fatalf("%v", err)
		return
	}

	opsCnt, err := reader.ReadUint32()
	if err != nil {
		t.Fatalf("%v", err)
		return
	}

	fmt.Printf("op results count: %d\n", opsCnt)

	for i := 0; i < int(opsCnt); i++ {
		opnum4, err := reader.ReadUint32()
		if err != nil {
			t.Fatalf("%v", err)
			return
		}

		opStatus, err := reader.ReadUint32()
		if err != nil {
			t.Fatalf("%v", err)
		}

		fmt.Printf(
			"op = %s, response status = %d.\n",
			nfs.Proc4Name(opnum4),
			opStatus,
		)

		switch opnum4 {
		case nfs.OP4_PUTFH:
			break

		case nfs.OP4_READDIR:
			cookieVerf := uint64(0)
			if _, err := reader.ReadAs(&cookieVerf); err != nil {
				t.Fatalf("%v", err)
				return
			}
			fmt.Printf(" - cookie_verf = %v\n", cookieVerf)

			hasEntries := false
			if _, err := reader.ReadAs(&hasEntries); err != nil {
				t.Fatalf("%v", err)
				return
			}
			fmt.Printf(" - has_entries = %v\n", hasEntries)

			entries := []*nfs.Entry4{}
			for {
				entry := &nfs.Entry4{}
				if _, err := reader.ReadAs(entry); err != nil {
					t.Fatalf("%v", err)
					return
				}
				entries = append(entries, entry)
				if !entry.HasNext {
					break
				}
			}

			for _, entry := range entries {
				fmt.Printf(" - entry: %s\n", entry.Name)
				fmt.Println(toJson(entry))
				if _, err := decodeFAttrs4(entry.Attrs); err != nil {
					t.Fatalf("%v", err)
				}
			}

			eof := false
			if _, err := reader.ReadAs(&eof); err != nil {
				t.Fatalf("%v", err)
				return
			}

			fmt.Printf(" - eof = %v\n", eof)

			break

		default:
			t.Fatalf("unexpected op: %s", nfs.Proc4Name(opnum4))
			return
		}
	}

}

func TestParsingCOMPOUND4_res_getAttr(t *testing.T) {
	raw, _ := base64.StdEncoding.DecodeString("KYbWCgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAGAAAAAAAAAAKAAAAAAAAAAgBAAEAAAAAAAAAAAkAAAAAAAAAAgAQARoAsKI6AAAAgAAAAAJgILVsOLsUNwAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANAABAAABwAAAAAQAAAABMAAAAAAAAAEwAAAAAAAAAAAAAAAAAAAAAAAQAAAAAABhlImSN76PiAAAAABgILVsOLsUNwAAAABgILVsOLsUNwAAAAAYMqia")

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

	// decode compound result...

	tag := ""
	if _, err := reader.ReadAs(&tag); err != nil {
		t.Fatalf("%v", err)
		return
	}

	opsCnt, err := reader.ReadUint32()
	if err != nil {
		t.Fatalf("%v", err)
		return
	}

	fmt.Printf("op results count: %d\n", opsCnt)

	for i := 0; i < int(opsCnt); i++ {
		opnum4, err := reader.ReadUint32()
		if err != nil {
			t.Fatalf("%v", err)
			return
		}

		opStatus, err := reader.ReadUint32()
		if err != nil {
			t.Fatalf("%v", err)
		}

		fmt.Printf(
			"op = %s, response status = %d.\n",
			nfs.Proc4Name(opnum4),
			opStatus,
		)

		switch opnum4 {
		case nfs.OP4_PUTROOTFH:
			break

		case nfs.OP4_GETFH:
			res := &nfs.GETFH4resok{}
			if _, err := reader.ReadAs(res); err != nil {
				t.Fatalf("%v", err)
				return
			}
			break

		case nfs.OP4_GETATTR:
			res := &nfs.GETATTR4resok{}
			if _, err := reader.ReadAs(res); err != nil {
				t.Fatalf("%v", err)
				return
			}

			fmt.Println(toJson(res))

			if _, err := decodeFAttrs4(res.Attr); err != nil {
				t.Fatalf("decodeAttrs: %v", err)
			}

			break

		default:
			t.Fatalf("unexpected op: %s", nfs.Proc4Name(opnum4))
			return
		}
	}

}
