package implv4

import (
	"github.com/smallfz/libnfs-go/nfs"
)

func closeFile(x nfs.RPCContext, args *nfs.CLOSE4args) (*nfs.CLOSE4res, error) {
	seqId := uint32(0)
	if args != nil && args.OpenStateId != nil {
		seqId = args.OpenStateId.SeqId
	}

	f := x.Stat().RemoveOpenedFile(seqId)
	if f == nil {
		return &nfs.CLOSE4res{Status: nfs.NFS4ERR_INVAL}, nil
	} else {
		f.File().Close()
	}

	res := &nfs.CLOSE4res{
		Status: nfs.NFS4_OK,
		Ok:     args.OpenStateId,
	}
	return res, nil
}
