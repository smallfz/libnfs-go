package implv3

import (
	// "fmt"
	"github.com/smallfz/libnfs-go/log"
	"github.com/smallfz/libnfs-go/nfs"
	"time"
	// "github.com/davecgh/go-xdr/xdr2"
)

func FsInfo(h *nfs.RPCMsgCall, ctx nfs.RPCContext) (int, error) {
	r, w := ctx.Reader(), ctx.Writer()

	log.Info("handling fsinfo.")
	sizeConsumed := 0

	fh3 := []byte{}
	if size, err := r.ReadAs(&fh3); err != nil {
		return 0, err
	} else {
		sizeConsumed += size
	}

	log.Infof("fsinfo: root = %s", string(fh3))

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
	rs := &nfs.FSINFO3resok{
		ObjAttrs: &nfs.PostOpAttr{
			AttributesFollow: true,
			Attributes: &nfs.FileAttrs{
				Type:  nfs.FTYPE_NF3DIR,
				Mode:  uint32(0755),
				Size:  1 << 63,
				Used:  0,
				ATime: nfs.MakeNfsTime(now),
				MTime: nfs.MakeNfsTime(now),
				CTime: nfs.MakeNfsTime(now),
			},
		},
		Rtmax:       1024 * 1024 * 4,
		Rtpref:      1024 * 1024 * 4,
		Rtmult:      1,
		Wtmax:       1024 * 1024 * 64,
		Wtpref:      1024 * 1024 * 64,
		Wtmult:      1,
		Dtpref:      0,
		MaxFileSize: 1024 * 1024 * 1024 * 4,
		TimeDelta:   nfs.NFSTime{Seconds: 1, NanoSeconds: 0},
		Properties:  nfs.FSF3_CANSETTIME,
	}

	if _, err := w.WriteAny(rs); err != nil {
		return sizeConsumed, err
	}

	return sizeConsumed, nil
}
