package filesystem

import (
	"io"
)

// FileSystem provides an abstraction for file system operations.
type FileSystem interface {
	FileExists(path string) (bool, error)
	WriteBytes(path string, data []byte) error
	Write(path string, r io.Reader) error
	ReadBytes(path string) ([]byte, error)
}
