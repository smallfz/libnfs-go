package implv3

import (
	"bytes"
	"fmt"
	fstools "libnfs-go/fs"
	"libnfs-go/log"
	"libnfs-go/nfs"
	"time"
)

func Lookup(h *nfs.RPCMsgCall, ctx nfs.RPCContext) (int, error) {
	r, w := ctx.Reader(), ctx.Writer()

	log.Info("handling lookup.")
	sizeConsumed := 0

	args := nfs.DirOpArgs3{}
	if size, err := r.ReadAs(&args); err != nil {
		return 0, err
	} else {
		sizeConsumed += size
	}

	log.Info(fmt.Sprintf(
		"lookup: dir = %v, filename = %v", args.Dir, args.Filename,
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

	// --- proc result ---

	zeros := string([]byte{0})
	folder := string(bytes.TrimRight(args.Dir, zeros))
	filename := args.Filename
	filename = fstools.Abs(fstools.Join(folder, filename))

	fs := ctx.GetFS()
	rsCode := nfs.NFS3_OK

	if fs != nil {

		di, err := fs.Stat(fstools.Abs(folder))
		if err != nil {
			if _, err := w.WriteUint32(nfs.NFS3ERR_ACCES); err != nil {
				return sizeConsumed, err
			}
			attributesFollow := false
			if _, err := w.WriteAny(attributesFollow); err != nil {
				return sizeConsumed, err
			}
			return sizeConsumed, nil
		}

		if fi, err := fs.Stat(filename); err == nil {

			if _, err := w.WriteUint32(nfs.NFS3_OK); err != nil {
				return sizeConsumed, err
			}

			now := time.Now()

			ftype := nfs.FTYPE_NF3DIR
			if !fi.IsDir() {
				ftype = nfs.FTYPE_NF3REG
			}

			attr := &nfs.LOOKUP3resok{
				Object: []byte(fi.Name()),
				ObjAttrs: &nfs.PostOpAttr{
					AttributesFollow: true,
					Attributes: &nfs.FileAttrs{
						Type:  ftype,
						Mode:  uint32(fi.Mode()),
						ATime: nfs.MakeNfsTime(now),
						MTime: nfs.MakeNfsTime(fi.ModTime()),
						CTime: nfs.MakeNfsTime(fi.ModTime()),
					},
				},
				DirAttrs: &nfs.PostOpAttr{
					AttributesFollow: true,
					Attributes: &nfs.FileAttrs{
						Type:  nfs.FTYPE_NF3DIR,
						Mode:  uint32(di.Mode()),
						ATime: nfs.MakeNfsTime(now),
						MTime: nfs.MakeNfsTime(di.ModTime()),
						CTime: nfs.MakeNfsTime(di.ModTime()),
					},
				},
			}
			if _, err := w.WriteAny(attr); err != nil {
				return sizeConsumed, err
			}
			return sizeConsumed, nil

		} else {
			log.Warn(fmt.Sprintf("fs.Stat(%s): %v", filename, err))
			rsCode = nfs.NFS3ERR_ACCES
		}
	} else {
		log.Warnf("no filesystem specified.")
		rsCode = nfs.NFS3ERR_IO
	}

	if _, err := w.WriteUint32(rsCode); err != nil {
		return sizeConsumed, err
	}

	// post_op_attr.attributes_follow
	attributesFollow := false
	if _, err := w.WriteAny(attributesFollow); err != nil {
		return sizeConsumed, err
	}

	return sizeConsumed, nil
}
