package implv4

import (
	"bytes"
	"fmt"
	"libnfs-go/fs"
	"libnfs-go/log"
	"libnfs-go/nfs"
	"libnfs-go/xdr"
	// "os"
	// "time"
)

const (
	A_supported_attrs    = 0  // bitmap4
	A_type               = 1  // nfs_ftype4, enum int
	A_fh_expire_type     = 2  // uint32
	A_change             = 3  // changeid4, uint64
	A_size               = 4  // uint64
	A_link_support       = 5  // bool
	A_symlink_support    = 6  // bool
	A_named_attr         = 7  // bool
	A_fsid               = 8  // fsid4, struct{uint64, uint64}
	A_unique_handles     = 9  // bool
	A_lease_time         = 10 // nfs_lease4, uint32
	A_rdattr_error       = 11 // nfsstat4, enum int
	A_acl                = 12 // nfsace4, struct{uint32, uint32, uint32, string}
	A_aclsupport         = 13 // uint32
	A_filehandle         = 19 // nfs_fh4, opaque<>
	A_fileid             = 20 // uint64
	A_mode               = 33 // (v4.1) uint32
	A_numlinks           = 35 // uint32
	A_owner              = 36 // string
	A_owner_group        = 37 // string
	A_rawdev             = 41 // struct{uint32, uint32}
	A_space_used         = 45 // uint64
	A_time_access        = 47 // nfstime4, struct{uint64, uint32}
	A_time_metadata      = 52 // nfstime4, struct{uint64, uint32}
	A_time_modify        = 53 // nfstime4, struct{uint64, uint32}
	A_mounted_on_fileid  = 55 // uint64
	A_suppattr_exclcreat = 75 // (v4.1) bitmap4
)

var (
	AttrsDefaultSet = []int{
		// A_supported_attrs,
		A_type,
		// A_fh_expire_type,
		A_change,
		A_size,
		// A_link_support,
		// A_symlink_support,
		// A_named_attr,
		A_fsid,
		// A_unique_handles,
		// A_lease_time,
		A_rdattr_error,
		A_filehandle,
		A_fileid,
		A_mode,
		A_numlinks,
		A_owner,
		A_owner_group,
		A_rawdev,
		A_space_used,
		A_time_access,
		A_time_metadata,
		A_time_modify,
		A_mounted_on_fileid,
		// A_suppattr_exclcreat,
	}
)

var (
	AttrsSupported = []int{
		A_supported_attrs,
		A_type,
		A_fh_expire_type,
		A_change,
		A_size,
		A_link_support,
		A_symlink_support,
		A_named_attr,
		A_fsid,
		A_unique_handles,
		A_lease_time,
		A_aclsupport,
		A_rdattr_error,
		A_filehandle,
		A_fileid,
		A_mode,
		A_numlinks,
		A_owner,
		A_owner_group,
		A_rawdev,
		A_space_used,
		A_time_access,
		A_time_metadata,
		A_time_modify,
		A_mounted_on_fileid,
		// A_suppattr_exclcreat,
	}
	AttrsWritable = []int{
		A_size,
		// A_acl,
		// A_archive,
		// A_hidden,
		// A_mimetype,
		A_mode,
		A_owner,
		A_owner_group,
		// A_system,
		// A_time_backup,
		// A_time_create,
	}
)

func GetAttrNameById(id int) (string, bool) {
	switch id {
	case A_supported_attrs:
		return "supported_attrs", true
	case A_type:
		return "type", true
	case A_fh_expire_type:
		return "fh_expire_type", true
	case A_change:
		return "change", true
	case A_size:
		return "size", true
	case A_link_support:
		return "link_support", true
	case A_symlink_support:
		return "symlink_support", true
	case A_named_attr:
		return "named_attr", true
	case A_fsid:
		return "fsid", true
	case A_unique_handles:
		return "unique_handles", true
	case A_lease_time:
		return "lease_time", true
	case A_rdattr_error:
		return "rdattr_error", true
	case A_filehandle:
		return "filehandle", true
	case A_fileid:
		return "fileid", true
	case A_mode:
		return "mode", true
	case A_numlinks:
		return "numlinks", true
	case A_owner:
		return "owner", true
	case A_owner_group:
		return "owner_group", true
	case A_rawdev:
		return "rawdev", true
	case A_space_used:
		return "space_used", true
	case A_time_access:
		return "time_access", true
	case A_time_metadata:
		return "time_metadata", true
	case A_time_modify:
		return "time_modify", true
	case A_mounted_on_fileid:
		return "mounted_on_fileid", true
	case A_suppattr_exclcreat:
		return "suppattr_exclcreat", true
	}
	return fmt.Sprintf("%d", id), false
}

func isAttrWritable(id int) bool {
	for _, a := range AttrsWritable {
		if a == id {
			return true
		}
	}
	return false
}

// getAttrsMaxBytesSize: estimate the byte size of a FAttr4.
func getAttrsMaxBytesSize(req map[int]bool) uint32 {
	idxSupport := map[int]bool{}
	for _, a := range AttrsSupported {
		idxSupport[a] = true
	}

	maxId := 0
	for a, _ := range req {
		if a > maxId {
			maxId = a
		}
	}

	idxSizes := map[int]int{
		A_supported_attrs:   4 * 2, // common case, estimated value
		A_type:              4,
		A_fh_expire_type:    4,
		A_change:            8,
		A_size:              8,
		A_link_support:      4,
		A_symlink_support:   4,
		A_named_attr:        4,
		A_fsid:              8 + 8,
		A_unique_handles:    4,
		A_lease_time:        4,
		A_aclsupport:        4,
		A_rdattr_error:      4,
		A_filehandle:        128 + 4, // max value
		A_fileid:            8,
		A_mode:              4,
		A_numlinks:          4,
		A_owner:             16, // estimated value
		A_owner_group:       16, // estimated value
		A_rawdev:            4 + 4,
		A_space_used:        8,
		A_time_access:       8 + 4,
		A_time_metadata:     8 + 4,
		A_time_modify:       8 + 4,
		A_mounted_on_fileid: 8,
	}

	total := uint32(0)
	for a := 0; a <= maxId; a++ {
		requested, found := req[a]
		if !found || !requested {
			continue
		}
		if ok, found := idxSupport[a]; !found || !ok {
			continue
		}
		if size, found := idxSizes[a]; found {
			total += uint32(size)
		}
	}

	return total
}

func fileInfoToAttrs(vfs fs.FS, pathName string, fi fs.FileInfo, attrsRequest map[int]bool) *nfs.FAttr4 {
	idxSupport := map[int]bool{}
	for _, a := range AttrsSupported {
		idxSupport[a] = true
	}

	idxDefault := map[int]bool{}
	for _, a := range AttrsDefaultSet {
		idxDefault[a] = true
	}

	if attrsRequest == nil {
		attrsRequest = idxDefault
	}

	maxId := 0
	for a, _ := range attrsRequest {
		if a > maxId {
			maxId = a
		}
	}

	idxReturn := map[int]bool{}
	for a, _ := range attrsRequest {
		idxReturn[a] = false
	}

	buff := bytes.NewBuffer([]byte{})
	w := xdr.NewWriter(buff)

	// debugf := func(a int, v interface{}) {
	// 	attrName, _ := GetAttrNameById(a)
	// 	switch a {
	// 	case A_type, A_mode, A_fileid, A_numlinks, A_mounted_on_fileid:
	// 		fallthrough
	// 	default:
	// 		log.Printf("     attr.%s = %v", attrName, v)
	// 		break
	// 	}
	// }

	writeAny := func(a int, target interface{}, sizeExpected int) {
		attrName, _ := GetAttrNameById(a)
		if size, err := w.WriteAny(target); err != nil {
			log.Errorf("attr.%s: w.WriteAny: %v", attrName, err)
		} else if size != sizeExpected {
			log.Errorf("attr.%s: w.WriteAny: %d bytes wrote(but expects %d).",
				attrName, size, sizeExpected,
			)
		} else {
			// debugf(a, target)
		}
	}

	for a := 0; a <= maxId; a++ {
		requested, found := attrsRequest[a]
		if !found || !requested {
			continue
		}

		supported, found := idxSupport[a]
		if !found || !supported {
			log.Warnf("attr requested but not supported: %d", a)
			continue
		}

		attrName, _ := GetAttrNameById(a)
		// log.Debugf(" - preparing attr: %s", attrName)

		idxReturn[a] = true
		switch a {
		case A_supported_attrs:
			v := bitmap4Encode(idxSupport)
			writeAny(a, v, 4+4*len(v))
			break
		case A_type:
			v := nfs.NF4REG
			if fi.IsDir() {
				v = nfs.NF4DIR
			}
			writeAny(a, v, 4)
			break
		case A_fh_expire_type:
			v := nfs.FH4_VOLATILE_ANY
			writeAny(a, v, 4)
			break
		case A_change:
			changeid := uint64(fi.ModTime().Unix())
			writeAny(a, changeid, 8)
			break
		case A_size:
			size := uint64(fi.Size())
			writeAny(a, size, 8)
			break
		case A_link_support:
			writeAny(a, true, 4)
			break
		case A_symlink_support:
			writeAny(a, false, 4)
			break
		case A_named_attr:
			writeAny(a, false, 4)
			break
		case A_fsid:
			fsid := &nfs.Fsid4{Major: 0, Minor: 0}
			writeAny(a, fsid, 8+8)
			break
		case A_unique_handles:
			writeAny(a, true, 4)
			break
		case A_lease_time:
			ttl := uint32(300) // seconds, rfc7530:5.8.1.11
			writeAny(a, ttl, 4)
			break
		case A_rdattr_error:
			status := nfs.NFS4_OK
			writeAny(a, status, 4)
			break
		case A_aclsupport:
			writeAny(a, uint32(0), 4)
			break
		case A_filehandle:
			fh, err := vfs.GetHandle(pathName)
			if err != nil {
				log.Warnf("vfs.GetHandle(%s): %v", pathName, err)
				if fh == nil {
					fh = []byte{}
				}
			}
			writeAny(a, fh, 4+len(fh)+xdr.Pad(len(fh)))
			break
		case A_fileid:
			fileid := vfs.GetFileId(pathName)
			writeAny(a, fileid, 8)
			break
		case A_mode:
			mask := (uint32(1) << 9) - 1
			mode := uint32(fi.Mode()) & mask
			writeAny(a, mode, 4)
			break
		case A_numlinks:
			n := uint32(0)
			switch i := fi.(type) {
			case fs.FileInfo:
				n = uint32(i.NumLinks())
				break
			}
			writeAny(a, n, 4)
			break
		case A_owner:
			writeAny(a, "0", 4+1+xdr.Pad(1))
			break
		case A_owner_group:
			writeAny(a, "0", 4+1+xdr.Pad(1))
			break
		case A_rawdev:
			v := &nfs.Specdata4{ /* uint32, uint32 */ }
			writeAny(a, v, 4+4)
			break
		case A_space_used:
			v := uint64(1024*4 + fi.Size())
			writeAny(a, v, 8)
			break
		case A_time_access:
			v := &nfs.NfsTime4{
				Seconds:  uint64(fi.ATime().Unix()),
				NSeconds: uint32(0),
			}
			writeAny(a, v, 8+4)
			break
		case A_time_metadata:
			v := &nfs.NfsTime4{
				Seconds: uint64(fi.CTime().Unix()),
			}
			writeAny(a, v, 8+4)
			break
		case A_time_modify:
			v := &nfs.NfsTime4{
				Seconds: uint64(fi.ModTime().Unix()),
			}
			writeAny(a, v, 8+4)
			break
		case A_mounted_on_fileid:
			fileid := vfs.GetFileId(pathName)
			writeAny(a, fileid, 8)
			break
		case A_suppattr_exclcreat:
			v := bitmap4Encode(idxSupport)
			writeAny(a, v, 4+4*len(v))
			break
		default:
			log.Warnf("(!)requested attr %s not handled!", attrName)
			break
		}
	}

	attrMask := bitmap4Encode(idxReturn)
	dat := buff.Bytes()

	return &nfs.FAttr4{
		Mask: attrMask,
		Vals: dat,
	}
}

type Attr struct {
	SupportedAttrs    []uint32 // bitmap4
	Type              uint32
	FhExpireType      uint32
	Change            uint64
	Size              *uint64
	LinkSupport       bool
	SymlinkSupport    bool
	NamedAttr         bool
	Fsid              *nfs.Fsid4
	UniqueHandles     bool
	LeaseTime         uint32
	RdattrError       uint32
	FileHandle        []byte
	FileId            uint64
	Mode              *uint32
	NumLinks          uint32
	Owner             string
	OwnerGroup        string
	Rawdev            *nfs.Specdata4
	SpaceUsed         uint64
	TimeAccess        *nfs.NfsTime4
	TimeMetadata      *nfs.NfsTime4
	TimeModify        *nfs.NfsTime4
	MountedOnFileId   uint64
	SuppAttrExclCreat []uint32
}

func decodeFAttrs4(attr *nfs.FAttr4) (*Attr, error) {
	decAttr := &Attr{}

	idx := nfs.Bitmap4Decode(attr.Mask)

	maxId := 0
	for aid, on := range idx {
		if on {
			if aid > maxId {
				maxId = aid
			}
		}
	}

	ar := xdr.NewReader(bytes.NewBuffer(attr.Vals))

	for i := 0; i <= maxId; i++ {
		if on, found := idx[i]; found && on {
			attrName, _ := GetAttrNameById(i)
			fmt.Printf(" - attr: %s\n", attrName)

			switch i {
			case A_supported_attrs:
				bm4 := []uint32{}
				if _, err := ar.ReadAs(&bm4); err != nil {
					return nil, err
				}
				decAttr.SupportedAttrs = bm4
				fmt.Printf("   value: %v\n", bm4)
				break
			case A_type:
				v := uint32(0)
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.Type = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_fh_expire_type:
				v := uint32(0)
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.FhExpireType = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_change:
				v := uint64(0)
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.Change = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_size:
				v := uint64(0)
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.Size = &v
				fmt.Printf("   value: %v\n", v)
				break
			case A_link_support:
				v := false
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.LinkSupport = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_symlink_support:
				v := false
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.SymlinkSupport = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_named_attr:
				v := false
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.NamedAttr = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_fsid:
				v := nfs.Fsid4{}
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.Fsid = &v
				fmt.Printf("   value: %v\n", v)
				break
			case A_unique_handles:
				v := false
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.UniqueHandles = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_lease_time:
				v := uint32(0)
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.LeaseTime = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_rdattr_error:
				v := uint32(0)
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.RdattrError = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_filehandle:
				v := []byte{}
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.FileHandle = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_fileid:
				v := uint64(0)
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.FileId = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_mode: // v4.1
				v := uint32(0)
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.Mode = &v
				fmt.Printf("   value: %v\n", v)
				break
			case A_numlinks:
				v := uint32(0)
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.NumLinks = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_owner:
				v := ""
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.Owner = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_owner_group:
				v := ""
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.OwnerGroup = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_rawdev:
				v := nfs.Specdata4{}
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.Rawdev = &v
				fmt.Printf("   value: %v\n", v)
				break
			case A_space_used:
				v := uint64(0)
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.SpaceUsed = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_time_access, A_time_metadata, A_time_modify:
				v := nfs.NfsTime4{}
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				switch i {
				case A_time_access:
					decAttr.TimeAccess = &v
					break
				case A_time_metadata:
					decAttr.TimeMetadata = &v
					break
				case A_time_modify:
					decAttr.TimeModify = &v
					break
				}
				fmt.Printf("   value: %v\n", v)
				break
			case A_mounted_on_fileid:
				v := uint64(0)
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.MountedOnFileId = v
				fmt.Printf("   value: %v\n", v)
				break
			case A_suppattr_exclcreat: // v4.1
				v := []uint32{}
				if _, err := ar.ReadAs(&v); err != nil {
					return nil, err
				}
				decAttr.SuppAttrExclCreat = v
				fmt.Printf("   value: %v\n", v)
				break
			}
		}
	}

	return decAttr, nil
}
