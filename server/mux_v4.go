package server

import (
	"fmt"

	"github.com/smallfz/libnfs-go/fs"
	"github.com/smallfz/libnfs-go/nfs"
	v4 "github.com/smallfz/libnfs-go/nfs/implv4"
	"github.com/smallfz/libnfs-go/xdr"
)

type Muxv4 struct {
	reader *xdr.Reader
	writer *xdr.Writer
	fs     fs.FS
	stat   nfs.StatService
}

func (x *Muxv4) Reader() *xdr.Reader {
	return x.reader
}

func (x *Muxv4) Writer() *xdr.Writer {
	return x.writer
}

func (x *Muxv4) Stat() nfs.StatService {
	return x.stat
}

func (x *Muxv4) GetFS() fs.FS {
	return x.fs
}

func (x *Muxv4) HandleProc(h *nfs.RPCMsgCall) (int, error) {
	switch h.Proc {
	case nfs.PROC4_VOID:
		return v4.Void(h, x)
	case nfs.PROC4_COMPOUND:
		return v4.Compound(h, x)
	}
	return 0, fmt.Errorf("not implemented: %s", nfs.Proc4Name(h.Proc))
}
