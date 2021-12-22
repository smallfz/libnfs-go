package implv4

import (
	"github.com/smallfz/libnfs-go/log"
	"github.com/smallfz/libnfs-go/nfs"
	"path"
)

func lookup(x nfs.RPCContext, args *nfs.LOOKUP4args) (*nfs.LOOKUP4res, error) {
	// log.Debugf("lookup obj: '%s'", args.ObjName)

	stat := x.Stat()
	pathName := path.Join(stat.Cwd(), args.ObjName)

	_, err := x.GetFS().Stat(pathName)
	if err != nil {
		log.Warnf(" lookup: %s: %v", pathName, err)
	}

	stat.SetCwd(pathName)

	res := &nfs.LOOKUP4res{
		Status: nfs.NFS4_OK,
	}
	return res, nil
}
