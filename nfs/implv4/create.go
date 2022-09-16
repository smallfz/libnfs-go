package implv4

import (
	"github.com/smallfz/libnfs-go/fs"
	"github.com/smallfz/libnfs-go/log"
	"github.com/smallfz/libnfs-go/nfs"
	"github.com/smallfz/libnfs-go/xdr"
	"os"
)

func readOpCreateArgs(r *xdr.Reader) (*nfs.CREATE4args, int, error) {
	sizeConsumed := 0
	typ, err := r.ReadUint32()
	if err != nil {
		return nil, sizeConsumed, err
	} else {
		sizeConsumed += 4
	}

	args := &nfs.CREATE4args{
		ObjType: typ,
	}

	switch typ {
	case nfs.NF4LNK:
		linkData := ""
		if size, err := r.ReadAs(&linkData); err != nil {
			return nil, sizeConsumed, err
		} else {
			sizeConsumed += size
		}
		args.LinkData = linkData
		break

	case nfs.NF4BLK, nfs.NF4CHR:
		devData := &nfs.Specdata4{}
		if size, err := r.ReadAs(devData); err != nil {
			return nil, sizeConsumed, err
		} else {
			sizeConsumed += size
		}
		args.DevData = devData
		break
	}

	objName := ""
	if size, err := r.ReadAs(&objName); err != nil {
		return nil, sizeConsumed, err
	} else {
		sizeConsumed += size
	}
	args.ObjName = objName

	attrs := &nfs.FAttr4{}
	if size, err := r.ReadAs(attrs); err != nil {
		return nil, sizeConsumed, err
	} else {
		sizeConsumed += size
	}
	args.CreateAttrs = attrs

	return args, sizeConsumed, nil
}

func create(x nfs.RPCContext, args *nfs.CREATE4args) (*nfs.CREATE4res, error) {
	switch args.ObjType {
	case nfs.NF4DIR, nfs.NF4REG:
		break
	case nfs.NF4LNK:
		fallthrough
	case nfs.NF4BLK, nfs.NF4CHR, nfs.NF4FIFO, nfs.NF4SOCK:
		return &nfs.CREATE4res{Status: nfs.NFS4ERR_PERM}, nil
	default:
		return &nfs.CREATE4res{Status: nfs.NFS4ERR_BADTYPE}, nil
	}

	resFailPerm := &nfs.CREATE4res{Status: nfs.NFS4ERR_PERM}
	resFail500 := &nfs.CREATE4res{Status: nfs.NFS4ERR_SERVERFAULT}

	// cwd := x.Stat().Cwd()
	vfs := x.GetFS()

	fh := x.Stat().CurrentHandle()
	cwd, err := vfs.ResolveHandle(fh)
	if err != nil {
		return resFailPerm, nil
	}

	fi, err := vfs.Stat(cwd)
	if err != nil {
		log.Debugf("    create: vfs.Stat(%s): %v", cwd, err)
		return resFail500, nil
	} else if !fi.IsDir() {
		return resFailPerm, nil
	}

	pathName := fs.Join(cwd, args.ObjName)
	log.Debugf("    create: %s", pathName)
	if _, err := vfs.Stat(pathName); err == nil {
		return &nfs.CREATE4res{Status: nfs.NFS4ERR_EXIST}, nil
	}

	cinfo := &nfs.ChangeInfo4{}
	attrSet := []uint32{}

	decAttrs, err := decodeFAttrs4(args.CreateAttrs)
	if err != nil {
		log.Warnf("create: decodeFAttrs: %v", err)
		return resFailPerm, nil
	}

	switch args.ObjType {
	case nfs.NF4DIR:

		// create a directory

		mod := os.FileMode(0755)
		if decAttrs.Mode != nil {
			mod = os.FileMode(*decAttrs.Mode)
		}
		mod = mod | os.ModeDir

		if err := vfs.MkdirAll(pathName, mod); err != nil {
			log.Warnf("create: vfs.MkdirAll(%s): %v", pathName, err)
			return resFailPerm, nil
		}
		if fi, err := vfs.Stat(pathName); err != nil {
			log.Warnf("create: vfs.Stat(%s): %v", pathName, err)
			return resFailPerm, nil
		} else {
			attr := fileInfoToAttrs(vfs, pathName, fi, nil)
			attrSet = attr.Mask

			// set current fh to the newly created one.
			if fh, err := vfs.GetHandle(fi); err != nil {
				return resFailPerm, nil
			} else {
				x.Stat().SetCurrentHandle(fh)
			}

			break
		}

	case nfs.NF4REG:

		// create a regular file

		mod := os.FileMode(0644)
		if decAttrs.Mode != nil {
			mod = os.FileMode(*decAttrs.Mode)
		}

		flag := os.O_CREATE | os.O_RDWR | os.O_TRUNC
		if f, err := vfs.OpenFile(pathName, flag, mod); err != nil {
			log.Warnf("create: vfs.OpenFile: %v", err)
			return resFailPerm, nil
		} else {
			defer f.Close()
			if fi, err := f.Stat(); err != nil {
				log.Warnf("create: f.Stat(): %v", err)
				return resFailPerm, nil
			} else {
				attr := fileInfoToAttrs(vfs, pathName, fi, nil)
				attrSet = attr.Mask

				// set current fh to the newly created one.
				if fh, err := vfs.GetHandle(fi); err != nil {
					return resFailPerm, nil
				} else {
					x.Stat().SetCurrentHandle(fh)
				}

				break
			}
		}

	}

	res := &nfs.CREATE4res{
		Status: nfs.NFS4_OK,
		Ok: &nfs.CREATE4resok{
			CInfo:   cinfo,
			AttrSet: attrSet,
		},
	}
	return res, nil
}
