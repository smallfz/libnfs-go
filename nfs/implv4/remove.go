package implv4

import (
	"github.com/smallfz/libnfs-go/log"
	"github.com/smallfz/libnfs-go/nfs"
	"path"
)

func remove(x nfs.RPCContext, args *nfs.REMOVE4args) (*nfs.REMOVE4res, error) {
	log.Debugf("remove obj: '%s'", args.Target)

	stat := x.Stat()
	vfs := x.GetFS()

	pathName := path.Join(stat.Cwd(), args.Target)

	_, err := vfs.Stat(pathName)
	if err != nil {
		log.Warnf("  remove: vfs.Stat(%s): %v", pathName, err)
	}

	if err := vfs.Remove(pathName); err != nil {
		log.Warnf("vfs.Remove(%s): %v", pathName, err)
		return &nfs.REMOVE4res{Status: nfs.NFS4ERR_PERM}, nil
	}

	res := &nfs.REMOVE4res{
		Status: nfs.NFS4_OK,
		Ok: &nfs.REMOVE4resok{
			CInfo: &nfs.ChangeInfo4{
				Atomic: true,
				Before: 0,
				After:  0,
			},
		},
	}
	return res, nil
}
