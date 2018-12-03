// package tap provides common interfaces for data exchange
package tap

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	// ErrNotFile is raised when a file object that does not point to a real file is accessed.
	ErrNotFile = errors.New("not a file")
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
	i, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var f *os.File
	if !i.IsDir() {
		f, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	}

	abs, _ := filepath.Abs(path)
	return &File{
		Path:      path,
		AbsPath:   abs,
		Name:      i.Name(),
		Size:      i.Size(),
		Mode:      i.Mode(),
		Directory: i.IsDir(),
		File:      f,
	}, nil
}

type File struct {
	Name      string
	Path      string
	AbsPath   string
	Size      int64
	Mode      os.FileMode
	Directory bool
	File      *os.File
}

func (f File) String() string {
	return f.Name
}

func (f *File) Read(b []byte) (int, error) {
	if f.File == nil {
		return 0, ErrNotFile
	}
	return f.File.Read(b)
}

func (f *File) Close() error {
	if f.File == nil {
		return ErrNotFile
	}
	return f.File.Close()
}
