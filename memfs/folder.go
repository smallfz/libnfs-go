package memfs

// import (
// 	"io"
// 	"os"
// )
//
// type memFolder struct {
// 	fi       *fileInfo
// 	children []os.FileInfo
// }
//
// func newMemFolder(fi *fileInfo, children []os.FileInfo) *memFolder {
// 	return &memFolder{
// 		fi:       fi,
// 		children: children,
// 	}
// }
//
// func (f *memFolder) Name() string {
// 	return f.fi.name
// }
//
// func (f *memFolder) Stat() (os.FileInfo, error) {
// 	return f.fi, nil
// }
//
// func (f *memFolder) Read(buff []byte) (int, error) {
// 	return 0, io.EOF
// }
//
// func (f *memFolder) Write(data []byte) (int, error) {
// 	return 0, io.EOF
// }
//
// func (f *memFolder) Seek(offset int64, whence int) (int64, error) {
// 	return 0, io.EOF
// }
//
// func (f *memFolder) Close() error {
// 	return nil
// }
//
// func (f *memFolder) Sync() error {
// 	return nil
// }
//
// func (f *memFolder) Readdir(n int) ([]os.FileInfo, error) {
// 	return f.children, nil
// }
//
