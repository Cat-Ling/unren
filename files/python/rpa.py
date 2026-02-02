import io, mmap, pickle, zlib

class GzipIO(io.IOBase):
    def __init__(self, file):
        self.file = file
        self.dc = zlib.decompressobj()
        self.buffer = bytearray()

    def readable(): return True
    def seekable(): return False
    def writable(): return False

    def read(self, n=-1):
        if n == -1:
            return self.readall()
        while n > len(self.buffer):
            b_in = bytearray(4096)
            count = self.file.readinto(b_in)
            self.buffer += b_in

            if count < 4096: # EOF
                self.buffer += self.dc.flush()
                break

        out = self.buffer[:n]
        del self.buffer[:n]
        return out
        pass

    def readall(self):
        pass

# Index format is a dictionary mapping file names to data blocks.
# Each block is specified as a 2- or 3-tuple of (offset, len, [start]).
# offset and len are obfuscated with the 32-bit key (xor cipher).
#
# Start appears to be used for buffering. In practice, it appears to be written
# to archives as an empty string. When reading a SubFile (loader.py:237), the
# start of a file from the index is prepended to the contents.
#
# Blocks are actually lists containing tuples; in practice, it seems the
# list is always only 1 element long.

class RPA(object):
    def __init__(self, fname):
        self.fname = fname
        with open(fname, "rb") as f:
            self.mm = mmap.mmap(f.fileno(), 0, access=mmap.ACCESS_READ)
        self._index = self.loadIndex()
        self.files = self.buildFileSet()

    def loadIndex(self):
        self.mm.seek(0)
        header = self.mm.readline()
        assert header.startswith(b"RPA-3.0 ")
            
        idx_ofs = int(header[8:24], 16)
        idx_key = int(header[25:33], 16)
        #with GzipIO(f) as gz:
        #    idx = pickle.load(gz)
        idx = pickle.loads(zlib.decompress(self.mm[idx_ofs:]))

        for k, v in idx.items():
            # Block is a list. Meaningful?
            assert len(v) == 1
            assert len(v[0]) in (2, 3)

            if len(v[0]) == 2:  # No start
                ofs, dlen = v[0]
                v[0] = (ofs ^ idx_key, dlen ^ idx_key, '')
            else:
                ofs, dlen, start = v[0]
                v[0] = (ofs ^ idx_key, dlen ^ idx_key, start)

        return idx

    def buildFileSet(self):
        fileset = {}
        for name, params in self._index.items():
            assert len(params) == 1
            ofs, dlen, start = params[0]
            assert not start
            fileset[name] = self.mm[ofs:ofs+dlen]
        return fileset

import os.path
def mkdirp(path):
    if len(path) == 0:
        return
    path = path.split('/')
    d = ''

    for i in range(len(path)):
        d = os.path.join(d, path[i])
        if not os.path.isdir(d):
            os.mkdir(d)


if __name__ == "__main__":
    import sys
    rpa = RPA(sys.argv[1])
    for file, contents in rpa.files.items():
        # print("{0:10} {1}".format(len(contents), file))
        mkdirp(os.path.dirname(file))
        with open(file, 'wb') as f:
            f.write(contents)
