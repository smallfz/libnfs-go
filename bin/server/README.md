

# How to


## Mount

### on macOS

    mount -o port=52590,mountport=52590,nfsvers=4,noacl,tcp -t nfs localhost:/ /Users/luoyun/mnt


### on linux

    mount -o port=52590,mountport=52590,nfsvers=4,minorversion=0,noacl,tcp -t nfs nfs-server:/ /mnt -v

