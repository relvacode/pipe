// package tap provides common interfaces for data exchange
package tap

import (
	"github.com/pkg/errors"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrDirectory = errors.New("cannot stream from a directory")
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

// Close closes x if x satisfies a Close() interface
func Close(x interface{}) error {
	c, ok := x.(closer)
	if ok {
		err := c.Close()
		LogError(err)
		return err
	}
	return nil
}

type forcedCloser struct {
	io.Reader
	orig io.Reader
}

func (c *forcedCloser) Close() error {
	LogError(Close(c.Reader))
	return Close(c.orig)
}

// ReadProxyCloser ensures that a wrapped reader stream is closed when the proxied reader is closed.
func ReadProxyCloser(wrapped, original io.Reader) io.ReadCloser {
	return &forcedCloser{
		Reader: wrapped,
		orig:   original,
	}
}

func OpenFileInfo(path string, i os.FileInfo) *File {
	var f = &File{
		Path:      path,
		Name:      i.Name(),
		Size:      i.Size(),
		Mode:      i.Mode(),
		Directory: i.IsDir(),
	}
	f.AbsPath, _ = filepath.Abs(path)
	if !f.Directory {
		f.Extension = filepath.Ext(path)
		f.Mime = mime.TypeByExtension(f.Extension)
	}
	return f
}

// OpenFile opens a file ready for reading.
func OpenFile(path string) (*File, error) {
	i, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return OpenFileInfo(path, i), nil
}

type File struct {
	Name      string
	Path      string
	AbsPath   string
	Size      int64
	Mode      os.FileMode
	Directory bool
	Extension string
	Mime      string

	f io.ReadCloser
}

func (f File) String() string {
	return f.Name
}

func (f *File) Read(b []byte) (int, error) {
	if f.Directory {
		return 0, ErrDirectory
	}
	if f.f == nil {
		var err error
		f.f, err = os.Open(f.Path)
		if err != nil {
			return 0, err
		}
	}
	return f.f.Read(b)
}

func (f *File) Close() error {
	if f.f == nil {
		return nil
	}
	return f.f.Close()
}
