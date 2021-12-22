// NFS server implements nfs v4.0
//
// Package server privodes a server implementation of nfs v4.
//
//     import (
//     	"fmt"
//     	"libnfs-go/log"
//     	"libnfs-go/memfs"
//     	"libnfs-go/server"
//     	"os"
//     )
//
//     mfs := memfs.NewMemFS()
//     backend := memfs.NewBackend(mfs)
//
//     mfs.MkdirAll("/mount", os.FileMode(0755))
//     mfs.MkdirAll("/test", os.FileMode(0755))
//     mfs.MkdirAll("/test2", os.FileMode(0755))
//     mfs.MkdirAll("/many", os.FileMode(0755))
//
//     perm := os.FileMode(0755)
//     for i := 0; i < 256; i++ {
//     	mfs.MkdirAll(fmt.Sprintf("/many/sub-%d", i+1), perm)
//     }
//
//     svr, err := server.NewServerTCP(2049, backend)
//     if err != nil {
//     	log.Errorf("server.NewServerTCP: %v", err)
//     	return
//     }
//
//     if err := svr.Serve(); err != nil {
//     	log.Errorf("svr.Serve: %v", err)
//     }
//
package server

import (
	"context"
	"fmt"
	"libnfs-go/log"
	"libnfs-go/nfs"
	"net"
)

type Server struct {
	listener net.Listener
	backend  nfs.Backend
}

func NewServerTCP(port int, backend nfs.Backend) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("net.Listen: %v", err)
	}
	return &Server{listener: ln, backend: backend}, nil
}

func (s *Server) Serve() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer s.listener.Close()

	log.Infof("Serving at %s ...", s.listener.Addr())

	for {
		if conn, err := s.listener.Accept(); err != nil {
			return fmt.Errorf("listener.Accept: %v", err)
		} else {
			go func() {
				defer conn.Close()
				if err := handleSession(ctx, s.backend, conn); err != nil {
					log.Errorf("%v", err)
				}
			}()
		}
	}

	return nil
}
