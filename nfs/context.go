package nfs

import (
	"github.com/smallfz/libnfs-go/fs"
	"github.com/smallfz/libnfs-go/xdr"
)

type RPCContext interface {
	Reader() *xdr.Reader
	Writer() *xdr.Writer
	GetFS() fs.FS
	Stat() StatService
}
