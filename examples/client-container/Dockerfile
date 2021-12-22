
FROM --platform=linux/amd64 centos:7

RUN [ -d /mnt ] || mkdir -p /mnt

RUN yum -y update
RUN yum install -y nfs-common rpcbind
RUN yum install -y python
# RUN yum install -y iproute2
# RUN yum install -y iptables
# RUN yum install -y iputils-ping
RUN yum install -y nfs-utils

RUN yum install -y python3
RUN yum install -y krb5-devel python3-devel swig python3-gssapi python3-ply
RUN yum install -y unzip
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


