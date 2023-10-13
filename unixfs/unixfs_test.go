package unixfs_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/smallfz/libnfs-go/fs"
	"github.com/smallfz/libnfs-go/unixfs"
)

var _ fs.FS = new(unixfs.UnixFS) // Check interface

func TestMemfsFileSeekRead(t *testing.T) {
	workdir, err := os.MkdirTemp("", "unixfs")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("workdir:", workdir)

	vfs, err := unixfs.New(workdir)
	if err != nil {
		t.Fatal(err)
	}

	//

	name := "hello.txt"
	content := "hello world!"
	pathName := fs.Join("/", name)

	perm := os.FileMode(0o644)
	flag := os.O_CREATE | os.O_RDWR

	if f, err := vfs.OpenFile(pathName, flag, perm); err != nil {
		t.Fatalf("OpenFile: %v", err)
		return
	} else {
		io.WriteString(f, content)
		f.Close()
	}

	if f, err := vfs.Open(pathName); err != nil {
		t.Fatalf("%v", err)
		return
	} else {
		offset := int64(len([]byte("hello ")))
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			t.Fatalf("%v", err)
			return
		}
		buff := bytes.NewBuffer([]byte{})
		if _, err := io.CopyN(buff, f, 1024); err != nil {
			if err != io.EOF {
				t.Fatalf("%v", err)
				return
			}
		}
		dat := buff.Bytes()
		if string(dat) != "world!" {
			t.Fatalf("unexpected content: %s", string(dat))
		}
		f.Close()
	}
}

func TestMemfsFileOpenTrunc(t *testing.T) {
	workdir, err := os.MkdirTemp("", "unixfs")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("workdir:", workdir)

	vfs, err := unixfs.New(workdir)
	if err != nil {
		t.Fatal(err)
	}

	//

	name := "hello.txt"
	content := "hello world!"
	pathName := fs.Join("/", name)

	perm := os.FileMode(0o644)
	flag := os.O_CREATE | os.O_RDWR

	if f, err := vfs.OpenFile(pathName, flag, perm); err != nil {
		t.Fatalf("OpenFile: %v", err)
		return
	} else {
		io.WriteString(f, content)
		f.Close()
	}

	if f, err := vfs.Open(pathName); err != nil {
		t.Fatalf("%v", err)
		return
	} else {
		dat, err := io.ReadAll(f)
		if err != nil {
			t.Fatalf("%v", err)
			return
		}
		if string(dat) != content {
			t.Fatalf("unexpected content: %s", string(dat))
		}
		f.Close()
	}

	// open truncate
	flag = os.O_RDWR | os.O_TRUNC
	if f, err := vfs.OpenFile(pathName, flag, perm); err != nil {
		t.Fatalf("%v", err)
		return
	} else {
		io.WriteString(f, "new content.")
		f.Close()
	}

	if f, err := vfs.Open(pathName); err != nil {
		t.Fatalf("%v", err)
		return
	} else {
		dat, err := io.ReadAll(f)
		if err != nil {
			t.Fatalf("%v", err)
			return
		}
		if string(dat) != "new content." {
			t.Fatalf("unexpected content: %s", string(dat))
		}
		f.Close()
	}
}

func TestMemfsFileOperations(t *testing.T) {
	workdir, err := os.MkdirTemp("", "unixfs")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("workdir:", workdir)

	vfs, err := unixfs.New(workdir)
	if err != nil {
		t.Fatal(err)
	}

	//

	folder := "/webapp"
	mode := os.FileMode(0o755)
	if err := vfs.MkdirAll(folder, mode); err != nil {
		t.Fatalf("MkdirAll(%s): %v", folder, err)
	}

	name := "hello.txt"
	content := "hello world!"

	pathName := fs.Join(folder, name)

	if _, err := vfs.Stat(pathName); err != nil {
		if !os.IsNotExist(err) {
			t.Fatalf("vfs.Stat: %v", err)
		} else {
			// should run into here..
		}
	} else {
		t.Fatalf("file should not be existing.")
	}

	// write file with data...

	mod := os.FileMode(0o644)
	if f, err := vfs.OpenFile(pathName, os.O_CREATE|os.O_RDWR, mod); err != nil {
		t.Fatalf("OpenFile: %v", err)
		return
	} else {
		io.WriteString(f, content)
		f.Close()
	}

	if fi, err := vfs.Stat(pathName); err != nil {
		t.Fatalf("(after created) vfs.Stat: %v", err)
	} else if fi.Name() != name {
		t.Fatalf("expects file-info with name `%s`. gets `%s`.",
			name, fi.Name(),
		)
	} else {
		// check file content...
		if f, err := vfs.Open(pathName); err != nil {
			t.Fatalf("vfs.Open: %v", err)
			return
		} else {
			if dat, err := io.ReadAll(f); err != nil {
				t.Fatalf("io.ReadAll: %v", err)
			} else {
				if string(dat) != content {
					t.Fatalf("wrong content: %v", string(dat))
				}
				fmt.Println(string(dat))
			}
			f.Close()
		}
	}

	// modify it's content

	contentNew := "great filesystem."
	flagWrite := os.O_RDWR | os.O_TRUNC

	if f, err := vfs.OpenFile(pathName, flagWrite, mod); err != nil {
		t.Fatalf("%v", err)
	} else {
		if _, err := io.WriteString(f, contentNew); err != nil {
			t.Fatalf("%v", err)
		}
		f.Close()
	}

	if f, err := vfs.Open(pathName); err != nil {
		t.Fatalf("%v", err)
	} else {
		if dat, err := io.ReadAll(f); err != nil {
			t.Fatalf("%v", err)
		} else if string(dat) != contentNew {
			t.Fatalf("failed to re-write file data. current content: %s",
				string(dat),
			)
		}
	}

	// check if the file exists in the expecting folder.

	checkExists := func(name string) {
		if d, err := vfs.Open("/webapp"); err != nil {
			t.Fatalf("%v", err)
		} else {
			if items, err := d.Readdir(-1); err != nil {
				t.Fatalf("%v", err)
			} else if len(items) != 1 {
				t.Fatalf("expects 1 file in /webapp. but gets %d.", len(items))
			} else {
				if items[0].Name() != name {
					t.Fatalf("the file is not the expected one: %s", items[0].Name())
				} else {
					fmt.Println(items[0].Name())
				}
			}
		}
	}

	checkExists(name)

	// rename

	newName := "new.txt"
	newPathName := fs.Join(folder, newName)
	if err := vfs.Rename(pathName, newPathName); err != nil {
		t.Fatalf("Rename: %v", err)
	}

	// check again

	checkExists(newName)

	// delete it

	if err := vfs.Remove(newPathName); err != nil {
		t.Fatalf("%v", err)
	}

	// check again: should be deleted

	if d, err := vfs.Open("/webapp"); err != nil {
		t.Fatalf("%v", err)
	} else {
		if items, err := d.Readdir(-1); err != nil {
			t.Fatalf("%v", err)
		} else if len(items) != 0 {
			t.Fatalf("expects no files in /webapp. but gets %d.", len(items))
		} else {
			fmt.Printf("%s: deleted as expected.\n", newName)
		}
	}
}

func TestMemfsMkdir(t *testing.T) {
	workdir, err := os.MkdirTemp("", "unixfs")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("workdir:", workdir)

	vfs, err := unixfs.New(workdir)
	if err != nil {
		t.Fatal(err)
	}

	// (1) create a folder

	pathName := "/abc/def/123"
	mode := os.FileMode(0o755)
	if err := vfs.MkdirAll(pathName, mode); err != nil {
		t.Fatalf("MkdirAll(%s): %v", pathName, err)
	}

	folder := fs.Dir(pathName)
	if folder != "/abc/def" {
		t.Fatalf("wrong dir returned from fs.Dir(%s): %s", pathName, folder)
	}

	f, err := vfs.Open(folder)
	if err != nil {
		t.Fatalf("%s: not exists.", folder)
	}

	// (2) create more dirs

	paths := []string{
		"/hello",
		"/data",
		"/data/backups",
		"/abc/www",
	}
	for _, pathName := range paths {
		mode := os.FileMode(0o755)
		if err := vfs.MkdirAll(pathName, mode); err != nil {
			t.Fatalf("MkdirAll(%s): %v", pathName, err)
		}
	}

	items, err := f.Readdir(-1)
	if err != nil {
		t.Fatalf("f.Readdir(-1): %v", err)
	}

	found := false
	for _, item := range items {
		fmt.Println(item.Name())
		if item.Name() == fs.Base(pathName) {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("target not created: %s", pathName)
		return
	}

	// (3) read dir of /abc

	d, err := vfs.Open("/abc")
	if err != nil {
		t.Fatalf("%v", err)
	}

	items, err = d.Readdir(-1)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expects 2 children of /abc. (gets %d)", len(items))
	}

	// (4) read dir of /

	d, err = vfs.Open("/")
	if err != nil {
		t.Fatalf("%v", err)
	}

	items, err = d.Readdir(-1)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(items) != 3 {
		t.Fatalf("expects 3 children of /. (gets %d)", len(items))
	}

	// (5) rename a dir
	if err := vfs.Rename("/data", "/data-new"); err != nil {
		t.Fatalf("Rename(/data => /data-new): %v", err)
	}

	if _, err := vfs.Open("/data"); err != nil {
		if !os.IsNotExist(err) {
			t.Fatalf("should not exists: /data (renamed to /data-new).")
			return
		}
	} else {
		t.Fatalf("should not exists: /data (renamed to /data-new).")
		return
	}

	if d, err := vfs.Open("/data-new"); err != nil {
		t.Fatalf("%v", err)
	} else {
		if items, err := d.Readdir(-1); err != nil {
			t.Fatalf("%v", err)
		} else {
			if len(items) != 1 {
				t.Fatalf("expects 1 children of /data-new. (gets %d)", len(items))
			}
		}
	}
}
