
import xdrlib
import socket
import struct


def main():
    addr = ('10.189.28.93', 29671)
    # addr = ('localhost', 29671)
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect(addr)
    try:
        pkr = xdrlib.Packer()
        xid = 0
        pkr.pack_uint(xid) # xid
        pkr.pack_uint(0) # rpc-type-call
        pkr.pack_uint(2) # rpc-ver
        pkr.pack_uint(100003) # nfs-program
        pkr.pack_uint(4) # nfs-ver

        # COMPOUND4args
        pkr.pack_string('readdir') # tag
        pkr.pack_uint(0) # minor-version
        pkr.pack_uint(1) # count of ops

        # op: readdir
        pkr.pack_uint(26) # OP_READDIR
        pkr.pack_uhyper(0) # cookie
        pkr.pack_uhyper(0) # cookie-verf
        pkr.pack_uint(32 * 1024)
        pkr.pack_uint(32 * 1024)

        # op: readdir: attr-request, bitmap4
        pkr.pack_uint(1)
        pkr.pack_uint(1 | (1<<1) | (1<<4) | (1<<7) | (1<<19))

        req = pkr.get_buffer()
        frag = (1 << 31) | len(req)
        frag = struct.pack('>I', frag)

        print(frag)
        sock.sendall(frag)
        sock.sendall(req)

        frag = struct.unpack('>I', sock.recv(4))[0]
        size = ((1 << 31) - 1) & frag
        print(bin(frag))
        print('response header size: %d' % size)
        dat = sock.recv(size)
        
        offset = 0
        while True:
            iv = dat[offset:offset+4]
            if not iv:
                break
            offset += len(iv)
            print(' '.join([bin(ord(c)) for c in iv]))
        # print('%x' % dat)
        # upkr = xdrlib.Unpacker(dat)
        # print('NFS4_OK: %d' % upkr.unpack_uint())
        # print(upkr.unpack_uint())
        # print(upkr.unpack_uint())
        # print(upkr.unpack_uint())
        # print(upkr.unpack_uint())
        # print(upkr.unpack_uint64())
    finally:
        if sock:
            sock.close()



if __name__ == '__main__':
    main()
    # pkr = xdrlib.Packer()
    # pkr.pack_string(b'123')
    # print(bytes(pkr.get_buffer()))

