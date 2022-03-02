package implv4

import (
	"bytes"
	"github.com/smallfz/libnfs-go/log"
	"github.com/smallfz/libnfs-go/nfs"
	"io"
	// "os"
)

func write(x nfs.RPCContext, args *nfs.WRITE4args) (*nfs.WRITE4res, error) {

	// stat := x.Stat()
	// vfs := x.GetFS()

	// pathName := stat.Cwd()

	// log.Debugf("write data to file: '%s'", pathName)
	// log.Printf(toJson(args))

	seqId := uint32(0)
	if args != nil && args.StateId != nil {
		seqId = args.StateId.SeqId
	}

	of := x.Stat().GetOpenedFile(seqId)
	if of == nil {
		return &nfs.WRITE4res{Status: nfs.NFS4ERR_INVAL}, nil
	}

	f := of.File()

	if args.Offset >= 0 {
		// log.Printf("  seek %d", args.Offset)
		if _, err := f.Seek(int64(args.Offset), io.SeekStart); err != nil {
			log.Warnf("f.Seek(%d): %v", args.Offset, err)
			return &nfs.WRITE4res{Status: nfs.NFS4ERR_PERM}, nil
		}
	}

	sizeWrote := uint32(0)
	if args.Data != nil && len(args.Data) > 0 {
		buff := bytes.NewReader(args.Data)
		if size, err := io.CopyN(f, buff, int64(len(args.Data))); err != nil {
			log.Warnf("io.CopyN(): %v", err)
			return &nfs.WRITE4res{Status: nfs.NFS4ERR_PERM}, nil
		} else {
			sizeWrote = uint32(size)
			// log.Printf("  %d bytes wrote.", sizeWrote)
		}
	} else {
		// log.Printf("  no data to be written.")
	}

	res := &nfs.WRITE4res{
		Status: nfs.NFS4_OK,
		Ok: &nfs.WRITE4resok{
			Count:     sizeWrote,
			Committed: nfs.FILE_SYNC4,
			WriteVerf: args.Offset,
		},
	}
	return res, nil
}
