package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"libnfs-go/log"
	"libnfs-go/nfs"
	"libnfs-go/nfs/implv4"
	"libnfs-go/xdr"
	"net"
	"sync"
)

const (
	localAddr  = ":29671"
	remoteAddr = "10.189.28.93:29671"
)

type nfsReqHead struct {
	RPCVer uint32 /* rfc1057, const: 2 */

	Prog uint32 /* nfs: 100003 */
	Vers uint32 /* 3 */
	Proc uint32 /* see proc.go */

	Cred *nfs.Auth
	Verf *nfs.Auth
}

func (h *nfsReqHead) String() string {
	procName := fmt.Sprintf("%d", h.Proc)
	if h.Prog == 100003 {
		switch h.Vers {
		case 3:
			procName = nfs.Proc3Name(h.Proc)
			break
		case 4:
			procName = nfs.Proc4Name(h.Proc)
			break
		}
	}
	return fmt.Sprintf(
		"<prog=%d, v=%d, proc=%s>",
		h.Prog, h.Vers, procName,
	)
}

func toJson(v interface{}) string {
	d, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(d)
}

type Session struct {
	currentProc uint32
}

func (s *Session) CurrentProc() uint32 {
	return s.currentProc
}

func (s *Session) SetProc(proc uint32) {
	s.currentProc = proc
}

func inspectNfs4CompoundProc(s *Session, r *xdr.Reader) {
	tag := ""
	if _, err := r.ReadAs(&tag); err != nil {
		return
	}

	minorVer := uint32(0)
	if _, err := r.ReadAs(&minorVer); err != nil {
		return
	}

	opsCnt := uint32(0)
	if _, err := r.ReadAs(&opsCnt); err != nil {
		return
	}

	for i := uint32(0); i < opsCnt; i++ {
		opnum4 := uint32(0)
		if _, err := r.ReadAs(&opnum4); err != nil {
			return
		}

		log.Infof(" op = %s", nfs.Proc4Name(opnum4))

		switch opnum4 {
		case nfs.OP4_SETCLIENTID:
			args := &nfs.SETCLIENTID4args{}
			if _, err := r.ReadAs(args); err != nil {
				return
			}
			log.Debugf(" args: %s", toJson(args))
			break

		case nfs.OP4_SETCLIENTID_CONFIRM:
			args := &nfs.SETCLIENTID_CONFIRM4args{}
			if _, err := r.ReadAs(args); err != nil {
				return
			}
			log.Debugf(" args: %s", toJson(args))
			break

		case nfs.OP4_EXCHANGE_ID:
			args := &nfs.EXCHANGE_ID4args{}
			if _, err := r.ReadAs(args); err != nil {
				return
			}
			log.Debugf(" args: %s", toJson(args))
			break

		case nfs.OP4_PUTROOTFH:
			break

		case nfs.OP4_GETATTR:
			args := &nfs.GETATTR4args{}
			if _, err := r.ReadAs(args); err != nil {
				return
			}
			log.Debugf(" args: %s", toJson(args))
			idx := nfs.Bitmap4Decode(args.AttrRequest)
			for aid, on := range idx {
				if on {
					name, _ := implv4.GetAttrNameById(aid)
					log.Debugf(" request attr: %s", name)
				}
			}
			break

		case nfs.OP4_PUTFH:
			args := &nfs.PUTFH4args{}
			if _, err := r.ReadAs(args); err != nil {
				return
			}
			log.Debugf(" args: %s", toJson(args))
			// log.Debugf(" args: %v", args)
			break

		case nfs.OP4_GETFH:
			break

		case nfs.OP4_LOOKUP:
			args := &nfs.LOOKUP4args{}
			if _, err := r.ReadAs(args); err != nil {
				return
			}
			log.Debugf(" args: %s", toJson(args))
			// log.Debugf(" args: %v", args)
			break

		case nfs.OP4_ACCESS:
			args := &nfs.ACCESS4args{}
			if _, err := r.ReadAs(args); err != nil {
				return
			}
			log.Debugf(" args: %s", toJson(args))
			// log.Debugf(" args: %v", args)
			break

		case nfs.OP4_READDIR:
			args := &nfs.READDIR4args{}
			if _, err := r.ReadAs(args); err != nil {
				return
			}
			log.Debugf(" args: %s", toJson(args))
			// log.Debugf(" args: %v", args)
			break

		case nfs.OP4_SECINFO:
			args := &nfs.SECINFO4args{}
			if _, err := r.ReadAs(args); err != nil {
				return
			}
			log.Debugf(" args: %s", toJson(args))
			// log.Debugf(" secinfo args.Name = %s", args.Name)
			break

		case nfs.OP4_RENEW:
			args := &nfs.RENEW4args{}
			if _, err := r.ReadAs(args); err != nil {
				return
			}
			break

		default:
			log.Warnf("op not handled: %d.", opnum4)
			return
		}

	}
}

func inspectNfs4Req(s *Session, h *nfsReqHead, reader *xdr.Reader) {
	s.SetProc(h.Proc)
	switch h.Proc {
	case nfs.PROC4_COMPOUND:
		inspectNfs4CompoundProc(s, reader)
		break
	default:
		log.Infof("proc %s. inspection skipped.", nfs.Proc4Name(h.Proc))
		break
	}
}

func inspectNfs4Response(s *Session, reader *xdr.Reader) {
	replyStat, err := reader.ReadUint32()
	if err != nil {
		return
	}

	if replyStat != nfs.ACCEPT_SUCCESS {
		log.Warnf(" reply-stat: %d", replyStat)
		return
	}

	auth := &nfs.Auth{}
	if _, err := reader.ReadAs(auth); err != nil {
		return
	}

	acceptStatus, err := reader.ReadUint32()
	if err != nil {
		return
	}

	if acceptStatus != nfs.ACCEPT_SUCCESS {
		log.Warnf(" accept-status: %d", acceptStatus)
		return
	}

	status, err := reader.ReadUint32()
	if err != nil {
		log.Warnf(" reader.ReadUint32() => status: %v", err)
		return
	}

	log.Printf(" status: %d", status)

	switch s.CurrentProc() {
	case nfs.PROC4_COMPOUND:
		lastStatus, err := reader.ReadUint32()
		if err != nil {
			return
		}
		log.Printf(" compound response: last status: %d", lastStatus)

		tag := ""
		if _, err := reader.ReadAs(&tag); err != nil {
			log.Warnf(" reader.ReadAs(&tag): %v", err)
			return
		}

		rsCount, err := reader.ReadUint32()
		if err != nil {
			return
		}

		log.Printf(" %d results of ops.", rsCount)

		for i := 0; i < int(rsCount); i++ {
			op, err := reader.ReadUint32()
			if err != nil {
				return
			}

			log.Printf(" op = %s", nfs.Proc4Name(op))

			opStatus, err := reader.ReadUint32()
			if err != nil {
				return
			}

			log.Printf("   > op result status: %d", opStatus)

			switch op {
			case nfs.OP4_SETCLIENTID:
				if opStatus == nfs.NFS4_OK {
					res := &nfs.SETCLIENTID4resok{}
					reader.ReadAs(res)
					log.Printf("  %s", toJson(res))
				} else if opStatus == nfs.NFS4ERR_CLID_INUSE {
					res := &nfs.ClientAddr4{}
					reader.ReadAs(res)
					log.Printf("  %s", toJson(res))
				}
				break

			case nfs.OP4_SETCLIENTID_CONFIRM:
				res := &nfs.SETCLIENTID_CONFIRM4res{}
				reader.ReadAs(res)
				log.Printf("  %s", toJson(res))
				break

			case nfs.OP4_EXCHANGE_ID:
				switch opStatus {
				case nfs.NFS4_OK:
					res := &nfs.EXCHANGE_ID4resok{}
					reader.ReadAs(res)
					log.Printf("  %s", toJson(res))
					break
				}
				break

			case nfs.OP4_PUTROOTFH:
				break

			case nfs.OP4_GETATTR:
				switch opStatus {
				case nfs.NFS4_OK:
					res := &nfs.GETATTR4resok{}
					reader.ReadAs(res)
					log.Printf("  %s", toJson(res))
					break
				}
				break

			case nfs.OP4_PUTFH:
				break

			case nfs.OP4_GETFH:
				switch opStatus {
				case nfs.NFS4_OK:
					res := &nfs.GETFH4resok{}
					reader.ReadAs(res)
					log.Printf("  %s", toJson(res))
					break
				}
				break

			case nfs.OP4_LOOKUP:
				break

			case nfs.OP4_ACCESS:
				switch opStatus {
				case nfs.NFS4_OK:
					res := &nfs.ACCESS4resok{}
					reader.ReadAs(res)
					log.Printf("  %s", toJson(res))
					break
				}
				break

			case nfs.OP4_READDIR:
				switch opStatus {
				case nfs.NFS4_OK:
					res := &nfs.READDIR4resok{}
					reader.ReadAs(res)
					log.Printf("  %s", toJson(res))
					break
				}
				break

			case nfs.OP4_SECINFO:
				break

			case nfs.OP4_RENEW:
				break

			default:
				break
			}
		}

		break
	}
}

func inspectNfsPacket(s *Session, tag string, dat []byte) {
	reader := xdr.NewReader(bytes.NewBuffer(dat))

	_, err := reader.ReadUint32() /* xid */
	if err != nil {
		return
	}
	msgType, err := reader.ReadUint32()
	if err != nil {
		return
	}

	switch msgType {
	case xdr.RPC_CALL:
		h := &nfsReqHead{}
		if _, err := reader.ReadAs(h); err != nil {
			log.Errorf("inspect(%s): ReadAs(h): %v", tag, err)
			return
		}
		log.Infof("inspect(%s): header: %v", tag, h)

		switch h.Vers {
		case 3:
			break
		case 4:
			inspectNfs4Req(s, h, reader)
			break
		}
		break

	case xdr.RPC_REPLY:
		// log.Printf("inspect: A reply received.")
		// inspectNfs4Response(s, reader)
		log.Printf(base64.StdEncoding.EncodeToString(dat))
		break

	default:
		break
	}
}

func proxyStream(s *Session, tag string, ctx context.Context, src io.Reader, dst io.Writer) error {
	reader := xdr.NewReader(src)
	writer := xdr.NewWriter(dst)

	for {
		frag, err := reader.ReadUint32()
		if err != nil {
			if err == io.EOF {
				return nil
			}
		}

		if frag&(1<<31) == 0 {
			return fmt.Errorf("(!)ignored: fragmented request")
		}

		headerSize := (frag << 1) >> 1
		restSize := int(headerSize)

		reqRawb, err := reader.ReadBytes(restSize)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("reader.ReadBytes: %v", err)
		}

		log.Infof("%s: packet size %d. %d bytes read.", tag, restSize, len(reqRawb))

		inspectNfsPacket(s, tag, reqRawb[:])

		if _, err := writer.WriteUint32(frag); err != nil {
			return fmt.Errorf("writer.WriteUint32(frag): %v", err)
		}

		if _, err := writer.Write(reqRawb); err != nil {
			return fmt.Errorf("writer.Write(%d bytes): %v", len(reqRawb), err)
		}
	}

	return nil
}

func handleSession(ctx context.Context, conn net.Conn) error {
	upConn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return fmt.Errorf("net.Dial: %v", err)
	}
	defer upConn.Close()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	s := &Session{}

	go func() {
		defer wg.Done()
		if err := proxyStream(s, "c=>s", ctx, conn, upConn); err != nil {
			log.Errorf("proxyStream(c => s): %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := proxyStream(s, "s=>c", ctx, upConn, conn); err != nil {
			log.Errorf("proxyStream(s => c): %v", err)
		}
	}()

	wg.Wait()

	return nil
}

func main() {
	ln, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Errorf("net.Listen: %v", err)
		return
	}
	defer ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Infof("Serving proxy at %s ...", ln.Addr())

	for {
		if conn, err := ln.Accept(); err != nil {
			log.Errorf("listener.Accept: %v", err)
			return
		} else {
			go func() {
				defer conn.Close()
				if err := handleSession(ctx, conn); err != nil {
					log.Errorf("%v", err)
				}
			}()
		}
	}

}
