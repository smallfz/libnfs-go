
[![Go Reference](https://pkg.go.dev/badge/github.com/smallfz/libnfs-go.svg)](https://pkg.go.dev/github.com/smallfz/libnfs-go)


# NFS Server Library

Experimental [NFSv4](https://datatracker.ietf.org/doc/html/rfc7530) server library in pure go.

The motive I started this project is mainly I need to immigrate a few projects to k8s and there are no other storage services except an OSS(like aws s3). So I need something to join OSS
and k8s PV together. To me NFS is a interesting choice.

Currently this repo doesn't include any COS wrapper in it.

## Getting start

To give it a try:

```go get github.com/smallfz/libnfs-go@main```

Demo memory fs:

```go
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/smallfz/libnfs-go/auth"
	"github.com/smallfz/libnfs-go/backend"
	"github.com/smallfz/libnfs-go/fs"
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

	// We don't need to create a new fs for each connection as memfs is opaque towards SetCreds.
	// If the file system would depend on SetCreds, make sure to generate a new fs.FS for each connection.
	backend := backend.New(func() fs.FS { return mfs }, auth.Null)

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
```

After the server started you can mount the in-memory filesystem to local:

on Mac:

```sh
mount -o nfsvers=4,noacl,tcp -t nfs localhost:/ /Users/smallfz/mnt
```

on Linux:

```sh
mount -o nfsvers=4,minorversion=0,noacl,tcp -t nfs localhost:/ /mnt
```


## Status

Recent testing results of [nfstest_posix](https://wiki.linux-nfs.org/wiki/index.php/NFStest) with `--nfsversion=4`:

 - access: 58/58 pass
 - open: 19/22 pass
 - chdir: 3/3 pass
 - readdir: 3/3 pass
 - opendir: 2/2 pass
 - seekdir,rewinddir: 9/9 pass
 - telldir: 5/5 pass
 - read: 3/3 pass
 - write: 7/7 pass
 - close: 3/3 pass
 - stat: 22/23 pass
 - fstat: 22/23 pass
 - chmod: 2/16 pass
 - creat: 6/6 pass
 - unlink: 2/4 pass
 - mkdir: 5/7 pass
 - rmdir: 3/5 pass
 - link: x
 - fcntl: x
 - ...

## Contributing

Firing an issue if anything. Any advice is appreciated.

