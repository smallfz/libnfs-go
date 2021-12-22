package implv3

import (
	"fmt"
	"libnfs-go/log"
	"libnfs-go/nfs"
	"time"
)

func FsStat(h *nfs.RPCMsgCall, ctx nfs.RPCContext) (int, error) {
	r, w := ctx.Reader(), ctx.Writer()

	log.Info("handling fsstat.")
	sizeConsumed := 0

	fh3 := []byte{} // nfs.Fh3{}
	if size, err := r.ReadAs(&fh3); err != nil {
		return 0, err
	} else {
		sizeConsumed += size
	}

	log.Info(fmt.Sprintf("fsstat: root = %v", fh3))

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

	// ---- proc result ---

	if _, err := w.WriteUint32(nfs.NFS3_OK); err != nil {
		return sizeConsumed, err
	}

	now := time.Now()
	capacity := uint64(1024 * 1024 * 1024 * 1024 * 2)

	rs := &nfs.FSSTAT3resok{
		ObjAttrs: &nfs.PostOpAttr{
			AttributesFollow: true,
			Attributes: &nfs.FileAttrs{
				Type:  nfs.FTYPE_NF3DIR,
				Mode:  uint32(0755),
				Size:  capacity,
				Used:  0,
				ATime: nfs.MakeNfsTime(now),
				MTime: nfs.MakeNfsTime(now),
				CTime: nfs.MakeNfsTime(now),
			},
		},
		Tbytes:   capacity,
		Fbytes:   capacity,
		Abytes:   capacity,
		Ffiles:   1024,
		Afiles:   1024,
		Invarsec: uint32(1),
	}

	if _, err := w.WriteAny(rs); err != nil {
		return sizeConsumed, err
	}

	return sizeConsumed, nil
}
