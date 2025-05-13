
FROM ubuntu:24.04

RUN [ -d /mnt ] || mkdir -p /mnt

RUN apt update
RUN apt install -y nfs-common rpcbind
RUN apt install -y nfs-kernel-server
# RUN apt install -y nfs-utils

RUN apt install -y python3
RUN apt install -y libkrb5-dev python3-dev swig python3-gssapi python3-ply
RUN apt install -y unzip
RUN pip3 install ply

ADD NFStest-1.0.10.tar.gz /opt
RUN cd /opt/NFStest-1.0.10 && python setup.py install

ADD pynfs-master.zip /opt
RUN cd /opt && unzip pynfs-master.zip
RUN cd /opt/pynfs-master && python3 setup.py install

RUN [ -d /run/sendsigs.omit.d ] || mkdir -p /run/sendsigs.omit.d

COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]

# mount -o port=2049,mountport=2049,nfsvers=4,minorversion=0,noacl,tcp -t nfs nfs-server:/ /mnt -v

# ./nfstest_posix -s nfs-server --port 2049 --nfsversion=4 --minorversion=0 -o "port=2049,mountport=2049,nfsvers=4,minorversion=0,noacl,tcp" -m /mnt -v debug --runtest open


