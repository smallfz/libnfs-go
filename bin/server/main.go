package main

import (
	"flag"
	"fmt"
	"libnfs-go/log"
	"libnfs-go/memfs"
	"libnfs-go/server"
	"os"
)

func main() {
	defer log.Sync()

	port := 2049
	flag.IntVar(&port, "p", port, "serving port.")
	flag.Parse()

	log.UpdateLevel(log.DEBUG)

	mfs := memfs.NewMemFS()
	backend := memfs.NewBackend(mfs)

	mfs.MkdirAll("/mount", os.FileMode(0755))
	mfs.MkdirAll("/test", os.FileMode(0755))
	mfs.MkdirAll("/test2", os.FileMode(0755))
	mfs.MkdirAll("/many", os.FileMode(0755))

	perm := os.FileMode(0755)
	for i := 0; i < 256; i++ {
		mfs.MkdirAll(fmt.Sprintf("/many/sub-%d", i+1), perm)
	}

	svr, err := server.NewServerTCP(port, backend)
	if err != nil {
		log.Errorf("server.NewServerTCP: %v", err)
		return
	}

	if err := svr.Serve(); err != nil {
		log.Errorf("svr.Serve: %v", err)
	}
}
