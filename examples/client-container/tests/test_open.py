
import os
import unittest


fpath = '/mnt/demo.dat'
content = '0123456789abcdef'


class TestCaseFileOpers(unittest.TestCase):

    def setUp(self):
        if os.path.isfile(fpath):
            os.unlink(fpath)
        f = open(fpath, 'wb')
        f.write(content)
        f.close()

    def tearDown(self):
        if os.path.isfile(fpath):
            os.unlink(fpath)

    def test_append(self):
        fd = os.open(fpath, os.O_APPEND|os.O_RDWR)
        os.write(fd, '____')
        os.close(fd)

        f = open(fpath, 'rb')
        dat = f.read()
        f.close()
        
        self.assertEqual(dat.decode('utf-8'), content + '____')

    def test_trunc(self):
        fd = os.open(fpath, os.O_TRUNC|os.O_RDWR)
        os.write(fd, '++++')
        os.close(fd)

        f = open(fpath, 'rb')
        dat = f.read()
        f.close()
        
        self.assertEqual(dat.decode('utf-8'), '++++')
        

if __name__ == '__main__':
    unittest.main()

