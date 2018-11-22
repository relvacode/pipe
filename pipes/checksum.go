package pipes

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"hash"
	"io"
)

var allHashGenerators = map[string]hashGenerator{
	"md5":    md5.New,
	"sha1":   sha1.New,
	"sha256": sha256.New,
	"sha512": sha512.New,
}

func init() {
	pkg := pipe.Pkg{
		Name: "checksum",
	}
	for k := range allHashGenerators {
		g := allHashGenerators[k]
		pkg.Family = append(pkg.Family, pipe.Pkg{
			Name: k,
			Constructor: func(command *console.Command) pipe.Pipe {
				return &ChecksumPipe{g: g}
			},
		})
	}
	pipe.Define(pkg)
}

type hashGenerator func() hash.Hash

type ChecksumPipe struct {
	g hashGenerator
}

func (p ChecksumPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		r, err := tap.Reader(f.Object)
		if err != nil {
			return err
		}

		h := p.g()
		_, err = io.Copy(h, r)
		if err != nil {
			return err
		}

		err = stream.Write(fmt.Sprintf("%x", h.Sum(nil)))
		if err != nil {
			return err
		}
	}
}
