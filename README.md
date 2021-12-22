
[![Go Reference](https://pkg.go.dev/badge/github.com/smallfz/libnfs-go.svg)](https://pkg.go.dev/github.com/smallfz/libnfs-go)


# NFS server library

Experimental NFS server library in pure go.

NFSv4 implementation posix tests:

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


Example:

see [bin/server/main.go](bin/server/main.go).

