package tap

import (
	"io"
	"io/ioutil"
	"os"
)

// MkTemp generates a temporary file containing x and returns the path for that file.
func MkTemp(x interface{}) (string, error) {
	r, err := Reader(x)
	if err != nil {
		return "", err
	}

	t, err := ioutil.TempFile(os.TempDir(), "pipe")
	if err != nil {
		return "", err
	}

	_, err = io.Copy(t, r)
	Close(r)
	if err != nil {
		t.Close()
		os.Remove(t.Name())
		return "", err
	}

	err = t.Close()
	if err != nil {
		os.Remove(t.Name())
		return "", err
	}

	Defer(func() error {
		return os.Remove(t.Name())
	})

	return t.Name(), nil
}
