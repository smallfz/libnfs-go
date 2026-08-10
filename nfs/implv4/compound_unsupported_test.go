package implv4

import (
	"bytes"
	"testing"

	"github.com/smallfz/libnfs-go/auth"
	"github.com/smallfz/libnfs-go/backend"
	"github.com/smallfz/libnfs-go/fs"
	"github.com/smallfz/libnfs-go/memfs"
	"github.com/smallfz/libnfs-go/nfs"
	"github.com/smallfz/libnfs-go/xdr"
)

type testRPCContext struct {
	reader *xdr.Reader
	writer *xdr.Writer
	vfs    fs.FS
	stat   nfs.StatService
}

var _ nfs.RPCContext = (*testRPCContext)(nil)

func (x *testRPCContext) Reader() *xdr.Reader { return x.reader }

func (x *testRPCContext) Writer() *xdr.Writer { return x.writer }

func (x *testRPCContext) Authenticate(_, _ *nfs.Auth) (*nfs.Auth, error) {
	return nfs.NewEmptyAuth(), nil
}

func (x *testRPCContext) GetFS() fs.FS { return x.vfs }

func (x *testRPCContext) Stat() nfs.StatService { return x.stat }

// callCompound runs Compound over a COMPOUND4args body carrying the given
// opnums, and returns a reader positioned on the reply.
func callCompound(t *testing.T, opnums ...uint32) *xdr.Reader {
	t.Helper()

	args := bytes.NewBuffer(nil)
	w := xdr.NewWriter(args)
	w.WriteAny("")                     /* tag */
	w.WriteUint32(0)                   /* minorversion */
	w.WriteUint32(uint32(len(opnums))) /* argarray size */
	for _, opnum4 := range opnums {
		w.WriteUint32(opnum4)
	}

	sess := backend.New(
		func() fs.FS { return memfs.NewMemFS() },
		auth.Null,
	).CreateSession(nil)

	reply := bytes.NewBuffer(nil)
	x := &testRPCContext{
		reader: xdr.NewReader(args),
		writer: xdr.NewWriter(reply),
		vfs:    sess.GetFS(),
		stat:   sess.GetStatService(),
	}

	h := &nfs.RPCMsgCall{
		Xid:     0x11223344,
		MsgType: nfs.RPC_CALL,
		Vers:    4,
		Proc:    nfs.PROC4_COMPOUND,
	}

	if _, err := Compound(h, x); err != nil {
		t.Fatalf("Compound: %v", err)
	}

	return xdr.NewReader(reply)
}

// readCompoundResHeader consumes the rpc reply header and returns the
// COMPOUND4res status.
func readCompoundResHeader(t *testing.T, reader *xdr.Reader) uint32 {
	t.Helper()

	if _, err := reader.ReadUint32(); err != nil { /* xid */
		t.Fatalf("%v", err)
	}

	msgType, err := reader.ReadUint32()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if msgType != xdr.RPC_REPLY {
		t.Fatalf("expects RPC_REPLY but get %d", msgType)
	}

	replyStat, err := reader.ReadUint32()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if replyStat != nfs.MSG_ACCEPTED {
		t.Fatalf(" reply-stat: %d", replyStat)
	}

	verf := &nfs.Auth{}
	if _, err := reader.ReadAs(verf); err != nil {
		t.Fatalf("%v", err)
	}

	acceptStatus, err := reader.ReadUint32()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if acceptStatus != nfs.ACCEPT_SUCCESS {
		t.Fatalf(" accept-status: %d", acceptStatus)
	}

	status, err := reader.ReadUint32()
	if err != nil {
		t.Fatalf(" reader.ReadUint32() => status: %v", err)
	}

	return status
}

// An op the server doesn't implement must still produce a decodable
// COMPOUND4res: rfc7530, 15.2. The op fails with NFS4ERR_NOTSUPP and the
// results of the ops before it are kept.
func TestCompound_unsupported_op(t *testing.T) {
	for _, tc := range []struct {
		name   string
		opnum4 uint32
		resOp  uint32
		status uint32
	}{
		{"lock", nfs.OP4_LOCK, nfs.OP4_LOCK, nfs.NFS4ERR_NOTSUPP},
		{"lockt", nfs.OP4_LOCKT, nfs.OP4_LOCKT, nfs.NFS4ERR_NOTSUPP},
		{"locku", nfs.OP4_LOCKU, nfs.OP4_LOCKU, nfs.NFS4ERR_NOTSUPP},
		{
			"release_lockowner",
			nfs.OP4_RELEASE_LOCKOWNER,
			nfs.OP4_RELEASE_LOCKOWNER,
			nfs.NFS4ERR_NOTSUPP,
		},
		// Not an op of the protocol at all: rfc7530, 15.2 wants this
		// reported against OP4_ILLEGAL.
		{"undefined", uint32(1024), nfs.OP4_ILLEGAL, nfs.NFS4ERR_OP_ILLEGAL},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := callCompound(t, nfs.OP4_PUTROOTFH, tc.opnum4)

			if status := readCompoundResHeader(t, reader); status != tc.status {
				t.Fatalf(" status: %d, expects %d", status, tc.status)
			}

			tag := ""
			if _, err := reader.ReadAs(&tag); err != nil {
				t.Fatalf("%v", err)
			}

			opsCnt, err := reader.ReadUint32()
			if err != nil {
				t.Fatalf("%v", err)
			}

			// putrootfh ran before the unsupported op. Its result must
			// still be there.
			if opsCnt != 2 {
				t.Fatalf("op results count: %d, expects 2", opsCnt)
			}

			results := []struct{ op, status uint32 }{}
			for i := uint32(0); i < opsCnt; i++ {
				op, err := reader.ReadUint32()
				if err != nil {
					t.Fatalf("%v", err)
				}

				opStatus, err := reader.ReadUint32()
				if err != nil {
					t.Fatalf("%v", err)
				}

				results = append(results, struct{ op, status uint32 }{op, opStatus})
			}

			if results[0].op != nfs.OP4_PUTROOTFH || results[0].status != nfs.NFS4_OK {
				t.Fatalf(
					"result 0: %s => %d, expects putrootfh => 0",
					nfs.Proc4Name(results[0].op),
					results[0].status,
				)
			}

			if results[1].op != tc.resOp || results[1].status != tc.status {
				t.Fatalf(
					"result 1: %s => %d, expects %s => %d",
					nfs.Proc4Name(results[1].op),
					results[1].status,
					nfs.Proc4Name(tc.resOp),
					tc.status,
				)
			}
		})
	}
}
