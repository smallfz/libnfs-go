package implv4

import (
	"io"
	"libnfs-go/fs"
	"libnfs-go/log"
	"libnfs-go/nfs"
)

func Compound(h *nfs.RPCMsgCall, ctx nfs.RPCContext) (int, error) {
	w := ctx.Writer()

	rh := &nfs.RPCMsgReply{
		Xid:       h.Xid,
		MsgType:   nfs.RPC_REPLY,
		ReplyStat: nfs.ACCEPT_SUCCESS,
	}
	if _, err := w.WriteAny(rh); err != nil {
		return 0, err
	}

	auth := &nfs.Auth{
		Flavor: nfs.AUTH_FLAVOR_NULL,
		Body:   []byte{},
	}
	if _, err := w.WriteAny(auth); err != nil {
		return 0, err
	}

	acceptStat := nfs.ACCEPT_SUCCESS
	if _, err := w.WriteUint32(acceptStat); err != nil {
		return 0, err
	}

	// ---- proc ----

	r, w := ctx.Reader(), ctx.Writer()
	sizeConsumed := 0

	tag := ""
	if size, err := r.ReadAs(&tag); err != nil {
		return sizeConsumed, err
	} else {
		sizeConsumed += size
	}

	minorVer := uint32(0)
	if size, err := r.ReadAs(&minorVer); err != nil {
		return sizeConsumed, err
	} else {
		sizeConsumed += size
	}

	opsCnt := uint32(0)
	if size, err := r.ReadAs(&opsCnt); err != nil {
		return sizeConsumed, err
	} else {
		sizeConsumed += size
	}

	log.Infof("---------- compound proc (%d ops) ----------", opsCnt)

	rsStatusList := []uint32{}
	rsOpList := []uint32{}
	rsList := []interface{}{}

	for i := uint32(0); i < opsCnt; i++ {
		opnum4 := uint32(0)
		if size, err := r.ReadAs(&opnum4); err != nil {
			return sizeConsumed, err
		} else {
			sizeConsumed += size
		}
		log.Infof("(%d) %s", i, nfs.Proc4Name(opnum4))
		switch opnum4 {
		case nfs.OP4_SETCLIENTID:
			args := &nfs.SETCLIENTID4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}
			res, err := setClientId(args)
			if err != nil {
				return sizeConsumed, err
			}
			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_SETCLIENTID_CONFIRM:
			args := &nfs.SETCLIENTID_CONFIRM4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}
			res, err := setClientIdConfirm(args)
			if err != nil {
				return sizeConsumed, err
			}
			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_EXCHANGE_ID:
			args := &nfs.EXCHANGE_ID4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			// todo: ...
			res := &nfs.EXCHANGE_ID4res{
				Status: nfs.NFS4_OK,
				Ok: &nfs.EXCHANGE_ID4resok{
					ClientId:     0,
					SequenceId:   0,
					StateProtect: &nfs.StateProtect4R{},
					ServerImplId: &nfs.NfsImplId4{
						Date: &nfs.NfsTime4{},
					},
				},
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_PUTROOTFH:
			// reset cwd to /
			stat := ctx.Stat()
			stat.SetCwd("/")

			// log.Infof("  putrootfh(/)")

			res := &nfs.PUTROOTFH4res{
				Status: nfs.NFS4_OK,
			}
			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_GETATTR:
			args := &nfs.GETATTR4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			res, err := getAttr(ctx, args)
			if err != nil {
				return sizeConsumed, err
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_PUTFH:
			args := &nfs.PUTFH4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			log.Debugf("    fh = %s", string(args.Fh))

			// set current handler to args.Fh
			stat := ctx.Stat()
			stat.SetCwd(fs.Abs(string(args.Fh)))
			// log.Debugf("  cwd set to %s", stat.Cwd())

			res := &nfs.PUTFH4res{
				Status: nfs.NFS4_OK,
			}
			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_GETFH:
			stat := ctx.Stat()
			vfs := ctx.GetFS()
			res := (*nfs.GETFH4res)(nil)
			fh, err := vfs.GetHandle(stat.Cwd())
			if err != nil {
				res = &nfs.GETFH4res{
					Status: nfs.NFS4ERR_SERVERFAULT,
				}
			} else {
				res = &nfs.GETFH4res{
					Status: nfs.NFS4_OK,
					Ok: &nfs.GETFH4resok{
						Fh: fh,
					},
				}
			}
			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_LOOKUP:
			args := &nfs.LOOKUP4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			res, err := lookup(ctx, args)
			if err != nil {
				log.Warnf("lookup: %v", err)
				return sizeConsumed, err
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_ACCESS:
			args := &nfs.ACCESS4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			res, err := access(ctx, args)
			if err != nil {
				log.Warnf("access: %v", err)
				return sizeConsumed, err
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_READDIR:
			args := &nfs.READDIR4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			res, err := readDir(ctx, args)
			if err != nil {
				log.Warnf("readdir: %v", err)
				return sizeConsumed, err
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_SECINFO:
			args := &nfs.SECINFO4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			// log.Debugf("secinfo args.Name = %s", args.Name)
			res := &nfs.SECINFO4res{
				Status: nfs.NFS4_OK,
				Ok: &nfs.SECINFO4resok{
					Items: []*nfs.Secinfo4{
						&nfs.Secinfo4{
							Flavor: 0,
							FlavorInfo: &nfs.RPCSecGssInfo{
								Service: nfs.RPC_GSS_SVC_NONE,
							},
						},
					},
				},
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_RENEW:
			args := &nfs.RENEW4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			// todo: renew client registration. somehow...

			res := &nfs.RENEW4res{
				Status: nfs.NFS4_OK,
			}
			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_CREATE:
			// rfc7530, 16.4.2
			args, size, err := readOpCreateArgs(r)
			if err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			res, err := create(ctx, args)
			if err != nil {
				return sizeConsumed, err
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_OPEN:
			args, size, err := readOpOpenArgs(r)
			if err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			res, err := open(ctx, args)
			if err != nil {
				return sizeConsumed, err
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_CLOSE:
			args := &nfs.CLOSE4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			res, err := closeFile(ctx, args)
			if err != nil {
				return sizeConsumed, err
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_SETATTR:
			args := &nfs.SETATTR4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			res, err := setAttr(ctx, args)
			if err != nil {
				return sizeConsumed, err
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_REMOVE:
			args := &nfs.REMOVE4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			res, err := remove(ctx, args)
			if err != nil {
				return sizeConsumed, err
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_WRITE:
			args := &nfs.WRITE4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			res, err := write(ctx, args)
			if err != nil {
				return sizeConsumed, err
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_READ:
			args := &nfs.READ4args{}
			if size, err := r.ReadAs(args); err != nil {
				return sizeConsumed, err
			} else {
				sizeConsumed += size
			}

			res, err := read(ctx, args)
			if err != nil {
				return sizeConsumed, err
			}

			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_SAVEFH:
			ctx.Stat().PushHandle(ctx.Stat().Cwd())
			res := &nfs.SAVEFH4res{
				Status: nfs.NFS4_OK,
			}
			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		case nfs.OP4_RESTOREFH:
			pathName, ok := ctx.Stat().PopHandle()
			if ok {
				ctx.Stat().SetCwd(pathName)
			}
			res := &nfs.RESTOREFH4res{
				Status: nfs.NFS4_OK,
			}
			rsOpList = append(rsOpList, opnum4)
			rsStatusList = append(rsStatusList, res.Status)
			rsList = append(rsList, res)
			break

		default:
			log.Warnf("op not handled: %d.", opnum4)
			w.WriteUint32(nfs.NFS4ERR_OP_ILLEGAL)
			return sizeConsumed, nil
		}
	}

	lastStatus := nfs.NFS4_OK
	if len(rsStatusList) > 0 {
		lastStatus = rsStatusList[len(rsStatusList)-1]
	}

	w.WriteUint32(lastStatus)
	w.WriteAny(tag) // tag: use the same as in request.

	w.WriteUint32(uint32(len(rsStatusList)))
	for i, rs := range rsList {
		op := rsOpList[i]

		w.WriteUint32(op)

		switch res := rs.(type) {
		case *nfs.ResGenericRaw:
			w.WriteUint32(res.Status)
			if res.Reader != nil {
				io.Copy(w, res.Reader)
			}
			break

		default:
			w.WriteAny(rs)
			break
		}
	}

	return sizeConsumed, nil
}