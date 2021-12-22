package implv3

import (
	"fmt"
	"libnfs-go/log"
	"libnfs-go/nfs"
	"time"
	// "github.com/davecgh/go-xdr/xdr2"
)

// GetAttr:
//
// SYNOPSIS
//
//     GETATTR3res NFSPROC3_GETATTR(GETATTR3args) = 1;
//
//     struct GETATTR3args {
//        nfs_fh3  object;
//     };
//
//     struct GETATTR3resok {
//        fattr3   obj_attributes;
//     };
//
//     union GETATTR3res switch (nfsstat3 status) {
//     case NFS3_OK:
//        GETATTR3resok  resok;
//     default:
//        void;
//     };
//
func GetAttr(h *nfs.RPCMsgCall, ctx nfs.RPCContext) (int, error) {
	r, w := ctx.Reader(), ctx.Writer()

	log.Info("handling getattr.")
	sizeConsumed := 0

	fh3 := []byte{}
	if size, err := r.ReadAs(&fh3); err != nil {
		return 0, err
	} else {
		sizeConsumed += size
	}

	log.Infof("getattr: fh3 = %s", string(fh3))

	rh := &nfs.RPCMsgReply{
		Xid:       h.Xid,
		MsgType:   nfs.RPC_REPLY,
		ReplyStat: nfs.ACCEPT_SUCCESS,
	}
	if _, err := w.WriteAny(rh); err != nil {
		return sizeConsumed, err
	}

	auth := &nfs.Auth{
		Flavor: nfs.AUTH_FLAVOR_NULL,
		Body:   []byte{},
	}
	if _, err := w.WriteAny(auth); err != nil {
		return sizeConsumed, err
	}

	acceptStat := nfs.ACCEPT_SUCCESS
	if _, err := w.WriteUint32(acceptStat); err != nil {
		return sizeConsumed, err
	}

	// --- proc result ---

	filename := string(fh3)
	fs := ctx.GetFS()
	rsCode := nfs.NFS3_OK

	if fs != nil {
		if fi, err := fs.Stat(filename); err == nil {

			if _, err := w.WriteUint32(nfs.NFS3_OK); err != nil {
				return sizeConsumed, err
			}

			now := time.Now()

			ftype := nfs.FTYPE_NF3DIR
			if !fi.IsDir() {
				ftype = nfs.FTYPE_NF3REG
			}

			attr := nfs.FileAttrs{
				Type:   ftype,
				Mode:   uint32(0755),
				Size:   uint64(fi.Size()),
				Used:   uint64(fi.Size()),
				Rdev:   nfs.SpecData{},
				Fsid:   0,
				FileId: 0,
				ATime:  nfs.MakeNfsTime(now),
				MTime:  nfs.MakeNfsTime(fi.ModTime()),
				CTime:  nfs.MakeNfsTime(fi.ModTime()),
			}
			if _, err := w.WriteAny(&attr); err != nil {
				return sizeConsumed, err
			}
			return sizeConsumed, nil

		} else {
			log.Warn(fmt.Sprintf("fs.Stat(%s): %v", filename, err))
			rsCode = nfs.NFS3ERR_BADHANDLE
		}
	} else {
		log.Warnf("no filesystem specified.")
		rsCode = nfs.NFS3ERR_IO
	}

	if _, err := w.WriteUint32(rsCode); err != nil {
		return sizeConsumed, err
	}

	return sizeConsumed, nil
}
