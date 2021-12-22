package implv3

import (
	"fmt"
	"github.com/smallfz/libnfs-go/log"
	"github.com/smallfz/libnfs-go/nfs"
	"time"
)

func Access(h *nfs.RPCMsgCall, ctx nfs.RPCContext) (int, error) {
	r, w := ctx.Reader(), ctx.Writer()

	log.Info("handling access.")
	sizeConsumed := 0

	fh3 := []byte{}
	access := uint32(0)
	if size, err := r.ReadAs(&fh3); err != nil {
		return 0, err
	} else {
		sizeConsumed += size
	}
	if size, err := r.ReadAs(&access); err != nil {
		return sizeConsumed, err
	} else {
		sizeConsumed += size
	}

	log.Info(fmt.Sprintf(
		"access: root = %s, access = %x", string(fh3), access,
	))

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
	rs := nfs.ACCESS3resok{
		ObjAttrs: &nfs.PostOpAttr{
			AttributesFollow: true,
			Attributes: &nfs.FileAttrs{
				Type:  nfs.FTYPE_NF3DIR,
				Mode:  uint32(0777),
				Size:  0,
				Used:  0,
				ATime: nfs.MakeNfsTime(now),
				MTime: nfs.MakeNfsTime(now),
				CTime: nfs.MakeNfsTime(now),
			},
		},
		Access: nfs.ACCESS3_READ | nfs.ACCESS3_LOOKUP | nfs.ACCESS3_MODIFY | nfs.ACCESS3_EXTEND | nfs.ACCESS3_DELETE | nfs.ACCESS3_EXECUTE,
	}

	if _, err := w.WriteAny(&rs); err != nil {
		return sizeConsumed, err
	}

	return sizeConsumed, nil
}
