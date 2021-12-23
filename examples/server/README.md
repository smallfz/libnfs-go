

# NFS Server Example

To launch a test server with a builtin in-memory filesystem:

    go run .


Then mount it:

on Mac:
    
    mount -o nfsvers=4,noacl,tcp -t nfs localhost:/ /Users/smallfz/mnt

on Linux:

    mount -o nfsvers=4,minorversion=0,noacl,tcp -t nfs localhost:/ /mnt

