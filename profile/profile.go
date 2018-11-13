package profile

import (
	"bufio"
	"github.com/minio/go-homedir"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// ProfileFile is the file name containing the user's pipe profile in that user's home directory.
	File = ".pipe_profile"
)

// Load loads the user's alias profile.
func Load() error {
	f, err := OpenProfile()
	if err != nil {
		return err
	}
	defer f.Close()

	alias, err := GetAlias(f)
	if err != nil {
		return err
	}

	return RegisterAlias(alias)
}

func OpenProfile() (io.ReadCloser, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(home, File)
	return os.OpenFile(path, os.O_CREATE|os.O_RDONLY, os.FileMode(0755))
}

func GetAlias(r io.Reader) (map[string]string, error) {
	var (
		s     = bufio.NewScanner(r)
		alias = make(map[string]string)
	)

	for s.Scan() {
		if strings.HasPrefix(s.Text(), "#") {
			continue
		}
		parts := strings.Split(s.Text(), "=")
		if len(parts) < 2 {
			return nil, errors.Errorf("Expected name=alias in %q", s.Text())
		}

		alias[parts[0]] = strings.Join(parts[1:], "=")
	}

	return alias, nil
}

func RegisterAlias(alias map[string]string) error {
	for k, cmd := range alias {

		pipes, err := pipe.Parse(strings.NewReader(cmd), pipe.Lib)
		if err != nil {
			return errors.Wrapf(err, "parse alias %q", k)
		}

		pipe.Define(pipe.Pkg{
			Name: k,
			Constructor: func(console *console.Command) pipe.Pipe {
				return pipe.SubPipe(pipes)
			},
		})
	}
	return nil
}
