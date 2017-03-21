package filesystem

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type LocalFileSystem struct {
	root string
}

type LocalFileSystemOption func(*LocalFileSystem) error

func Root(path string) LocalFileSystemOption {
	if !strings.HasSuffix(path, string(os.PathSeparator)) {
		path += string(os.PathSeparator)
	}

	return func(fs *LocalFileSystem) error {

		fs.root = path

		return nil
	}
}

// NewLocalFileSystem creates a new local file system with the specified options.
func NewLocalFileSystem(options ...LocalFileSystemOption) (*LocalFileSystem, error) {
	fs := &LocalFileSystem{}

	for _, o := range options {
		if err := o(fs); err != nil {
			return nil, err
		}
	}

	return fs, nil
}

func (fs *LocalFileSystem) FileExists(path string) (bool, error) {
	fi, err := os.Stat(fs.root + path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	if fi.IsDir() {
		return false, fmt.Errorf("exists, but is a directory: %s", path)
	}

	return true, nil
}

func (fs *LocalFileSystem) WriteBytes(path string, data []byte) error {
	return ioutil.WriteFile(fs.root+path, data, os.ModePerm)
}

func (fs *LocalFileSystem) Write(path string, r io.Reader) error {
	f, err := os.Create(fs.root + path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)

	return err
}

func (fs *LocalFileSystem) ReadBytes(path string) ([]byte, error) {
	return ioutil.ReadFile(fs.root + path)
}
