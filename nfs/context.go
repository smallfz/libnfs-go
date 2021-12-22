package nfs

import (
	"libnfs-go/fs"
	"libnfs-go/xdr"
)

type RPCContext interface {
	Reader() *xdr.Reader
	Writer() *xdr.Writer
	GetFS() fs.FS
	Stat() StatService
}
