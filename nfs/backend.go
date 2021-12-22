package nfs

import (
	"libnfs-go/fs"
	"net"
)

// SessionState represents a client network session.
type SessionState interface {
	// Conn returns the current client connection.
	Conn() net.Conn
}

type StatService interface {
	Cwd() string
	SetCwd(string) error

	PushHandle(string)
	PopHandle() (string, bool)

	SetClientId(uint64)
	ClientId() (uint64, bool)

	AddOpenedFile(string, fs.File) uint32
	GetOpenedFile(uint32) fs.FileOpenState
	RemoveOpenedFile(uint32) fs.FileOpenState

	// CleanUp should remove all opened files and reset handle stack.
	CleanUp()
}

// BackendSession has a lifetime exact as the client connection.
type BackendSession interface {

	// GetFS should return a FS implementation.
	GetFS() fs.FS

	// GetStatService returns a StateService in implementation.
	// In development you can return a memfs.Stat instance.
	GetStatService() StatService

	// Close invoked by server when connection closed by any side.
	// Implementation should do some cleaning work at this time.
	Close() error
}

// Backend interface. This is where it starts when building a custom nfs server.
type Backend interface {
	// CreateSession returns a session instance.
	// In development you can return a memfs.Backend instance.
	CreateSession(SessionState) BackendSession
}
