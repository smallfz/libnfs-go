package implv4

import (
	// "bytes"
	"libnfs-go/log"
	"libnfs-go/nfs"
	// "libnfs-go/xdr"
	// "encoding/base64"
)

func getAttr(x nfs.RPCContext, args *nfs.GETATTR4args) (*nfs.GETATTR4res, error) {
	cwd := x.Stat().Cwd()
	pathName := cwd
	log.Debugf("    getattr(%s)", pathName)

	idxReq := bitmap4Decode(args.AttrRequest)

	// log.Debugf("getattr: attrs: %v", idxReq)
	// for id, on := range idxReq {
	// 	if on {
	// 		name, found := GetAttrNameById(id)
	// 		if found {
	// 			log.Debugf(" - request attr: [%d] %s.", id, name)
	// 		}
	// 	}
	// }

	vfs := x.GetFS()
	fi, err := vfs.Stat(pathName)
	if err != nil {
		log.Warnf("vfs.Stat(%s): %v", cwd, err)
		return &nfs.GETATTR4res{Status: nfs.NFS4ERR_NOENT}, nil
	}

	_, err = vfs.GetHandle(pathName)
	if err != nil {
		log.Warnf("vfs.GetHandle: %v", err)
		return &nfs.GETATTR4res{Status: nfs.NFS4ERR_NOENT}, nil
	}

	attrs := fileInfoToAttrs(vfs, cwd, fi, idxReq)

	rs := &nfs.GETATTR4res{
		Status: nfs.NFS4_OK,
		Ok: &nfs.GETATTR4resok{
			Attr: attrs,
		},
	}
	return rs, nil
}
