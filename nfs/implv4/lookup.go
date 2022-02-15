package implv4

import (
	"github.com/smallfz/libnfs-go/log"
	"github.com/smallfz/libnfs-go/nfs"
	"path"
)

func lookup(x nfs.RPCContext, args *nfs.LOOKUP4args) (*nfs.LOOKUP4res, error) {
	// log.Debugf("lookup obj: '%s'", args.ObjName)

	if len(args.ObjName) <= 0 {
		return &nfs.LOOKUP4res{
			Status: nfs.NFS4ERR_INVAL,
		}, nil
	}

	stat := x.Stat()
	vfs := x.GetFS()

	fh := stat.CurrentHandle()
	folder, err := vfs.ResolveHandle(fh)
	if err != nil {
		log.Warnf("ResolveHandle: %v", err)
		return &nfs.LOOKUP4res{Status: nfs.NFS4ERR_PERM}, nil
	}

	pathName := path.Join(folder, args.ObjName)

	if fi, err := x.GetFS().Stat(pathName); err != nil {
		log.Warnf(" lookup: %s: %v", pathName, err)
		return &nfs.LOOKUP4res{
			Status: nfs.NFS4ERR_NOENT,
		}, nil
	} else {
		// stat.SetCwd(pathName)
		if fh, err := vfs.GetHandle(fi); err != nil {
			return &nfs.LOOKUP4res{
				Status: nfs.NFS4ERR_NOENT,
			}, nil
		} else {
			stat.SetCurrentHandle(fh)
		}
	}

	res := &nfs.LOOKUP4res{
		Status: nfs.NFS4_OK,
	}
	return res, nil
}
