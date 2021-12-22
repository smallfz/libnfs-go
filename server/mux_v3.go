package server

import (
	// "libnfs-go/log"
	"fmt"
	"libnfs-go/fs"
	"libnfs-go/nfs"
	handlers "libnfs-go/nfs/implv3"
	"libnfs-go/xdr"
)

type Mux struct {
	reader *xdr.Reader
	writer *xdr.Writer
	fs     fs.FS
	stat   nfs.StatService
}

func (x *Mux) Reader() *xdr.Reader {
	return x.reader
}

func (x *Mux) Writer() *xdr.Writer {
	return x.writer
}

func (x *Mux) Stat() nfs.StatService {
	return x.stat
}

func (x *Mux) GetFS() fs.FS {
	return x.fs
}

func (x *Mux) HandleProc(h *nfs.RPCMsgCall) (int, error) {
	switch h.Proc {
	case nfs.ProcVoid:
		return handlers.Void(h, x)
	case nfs.ProcGetAttr:
		return handlers.GetAttr(h, x)
	case nfs.ProcFsInfo:
		return handlers.FsInfo(h, x)
	case nfs.ProcPathConf:
		return handlers.PathConf(h, x)
	case nfs.ProcFsStat:
		return handlers.FsStat(h, x)
	case nfs.ProcAccess:
		return handlers.Access(h, x)
	case nfs.ProcLookup:
		return handlers.Lookup(h, x)
	case nfs.ProcReaddirPlus:
		return handlers.ReaddirPlus(h, x)
	}
	return 0, fmt.Errorf("not implemented.")
}