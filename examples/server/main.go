package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/smallfz/libnfs-go/backend"
	"github.com/smallfz/libnfs-go/log"
	"github.com/smallfz/libnfs-go/memfs"
	"github.com/smallfz/libnfs-go/server"
)

func main() {
	listen := ":2049"
	flag.StringVar(&listen, "l", listen, "Server listen address")
	flag.Parse()

	log.UpdateLevel(log.DEBUG)

	mfs := memfs.NewMemFS()
	backend := backend.New(mfs)

	mfs.MkdirAll("/mount", os.FileMode(0o755))
	mfs.MkdirAll("/test", os.FileMode(0o755))
	mfs.MkdirAll("/test2", os.FileMode(0o755))
	mfs.MkdirAll("/many", os.FileMode(0o755))

	perm := os.FileMode(0o755)
	for i := 0; i < 256; i++ {
		mfs.MkdirAll(fmt.Sprintf("/many/sub-%d", i+1), perm)
	}

	svr, err := server.NewServerTCP(listen, backend)
	if err != nil {
		log.Errorf("server.NewServerTCP: %v", err)
		return
	}

	if err := svr.Serve(); err != nil {
		log.Errorf("svr.Serve: %v", err)
	}
}
