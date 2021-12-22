package implv4

import (
	"io"
	"libnfs-go/fs"
	"libnfs-go/log"
	"libnfs-go/nfs"
	"os"
)

func setAttr(x nfs.RPCContext, args *nfs.SETATTR4args) (*nfs.SETATTR4res, error) {
	resFailNotSupp := &nfs.SETATTR4res{Status: nfs.NFS4ERR_ATTRNOTSUPP}
	resFailPerm := &nfs.SETATTR4res{Status: nfs.NFS4ERR_PERM}

	a4 := args.Attrs
	idxReq := bitmap4Decode(a4.Mask)

	// uncheck not-writable attributes

	off := map[int]bool{}
	for id, on := range idxReq {
		if on && !isAttrWritable(id) {
			off[id] = false
		}
	}
	for id, _ := range off {
		if on, found := idxReq[id]; found && on {
			idxReq[id] = false
		}
	}

	cwd := x.Stat().Cwd()
	vfs := x.GetFS()

	seqId := uint32(0)
	if args.StateId != nil {
		seqId = args.StateId.SeqId
	}

	f := (fs.File)(nil)
	pathName := cwd

	of := x.Stat().GetOpenedFile(seqId)

	if of != nil {
		f = of.File()
		pathName = of.Path()
	} else {
		if _f, err := vfs.Open(pathName); err != nil {
			log.Warnf("vfs.Open(%s): %v", pathName, err)
			return resFailPerm, nil
		} else {
			defer _f.Close()
			f = _f
		}
	}

	fi, err := f.Stat()
	if err != nil {
		log.Warnf("f.Stat: %v", err)
		return resFailPerm, nil
	}

	decAttrs, err := decodeFAttrs4(args.Attrs)
	if err != nil {
		return resFailNotSupp, nil
	}
	// log.Println(toJson(decAttrs))

	// TODO: actually set the attributes....
	if decAttrs.Mode != nil {
		perm := os.FileMode(*decAttrs.Mode)
		if err := vfs.Chmod(pathName, perm); err != nil {
			log.Warnf("vfs.Chmod(%s, %o): %v", pathName, perm, err)
			return resFailPerm, nil
		}
	}
	if decAttrs.Size != nil {
		size := int64(*decAttrs.Size)

		if _, err := f.Seek(size, io.SeekStart); err != nil {
			log.Warnf("f.Seek(%d, %d): %v", size, io.SeekStart, err)
		} else {
			if err := f.Truncate(); err != nil {
				log.Warnf("f.Truncate: %v", err)
				return resFailPerm, nil
			}
		}
	}

	fi, err = f.Stat()
	if err != nil {
		log.Warnf("f.Stat: %v", err)
		return resFailPerm, nil
	}

	attrs := fileInfoToAttrs(vfs, cwd, fi, idxReq)
	attrSet := attrs.Mask

	return &nfs.SETATTR4res{
		Status:  nfs.NFS4_OK,
		AttrSet: attrSet,
	}, nil
}
