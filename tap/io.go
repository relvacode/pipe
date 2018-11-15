// package tap provides common interfaces for data exchange
package tap

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
)

// Reader gets a reader interface from an input interface
func Reader(x interface{}) (io.Reader, error) {
	switch v := x.(type) {
	case io.Reader:
		return v, nil
	case string:
		return strings.NewReader(v), nil
	}
	return nil, errors.Errorf("Expected a string or file-like object but got type %T", x)
}

type closer interface {
	Close() error
}

func Close(x interface{}) error {
	c, ok := x.(closer)
	if ok {
		return c.Close()
	}
	return nil
}

type forcedCloser struct {
	io.Reader
	orig io.Reader
}

func (c *forcedCloser) Close() error {
	Close(c.Reader)
	return Close(c.orig)
}

// ReadProxyCloser ensures that a wrapped reader stream is closed when the proxied reader is closed.
func ReadProxyCloser(wrapped, original io.Reader) io.ReadCloser {
	return &forcedCloser{
		Reader: wrapped,
		orig:   original,
	}
}

// OpenFile opens a file ready for reading.
func OpenFile(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	i, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if i.IsDir() {
		return nil, errors.Errorf("cannot open %q: is directory", path)
	}

	return &File{
		Path: path,
		Name: f.Name(),
		Size: i.Size(),
		File: f,
	}, nil
}

type File struct {
	Name string
	Path string
	Size int64
	File *os.File
}

func (f *File) Read(b []byte) (int, error) {
	return f.File.Read(b)
}

func (f *File) Close() error {
	return f.File.Close()
}
