package pipes

import (
	"context"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"io"
	"net/http"
	"strings"
)

var methods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodDelete,
	http.MethodHead,
	http.MethodOptions,
	http.MethodPatch,
}

func init() {
	pkg := pipe.Pkg{
		Name: "url",
	}
	for i := 0; i < len(methods); i++ {
		method := methods[i]
		pkg.Family = append(pkg.Family, pipe.Pkg{
			Name: strings.ToLower(method),
			Constructor: func(console *console.Command) pipe.Pipe {
				cli := console.Options()
				return &URLPipe{
					method:  method,
					headers: cli.Option("header").Map(),
					url:     cli.Arg(0).String(),
				}
			},
		})
	}
	pipe.Define(pkg)
}

var _ io.ReadCloser = (*Response)(nil)

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       io.ReadCloser
}

func (r *Response) Read(b []byte) (int, error) {
	return r.Body.Read(b)
}

func (r *Response) Close() error {
	return r.Body.Close()
}

// URLPipe calls a URL and emits a Response to the stream
type URLPipe struct {
	method  string
	headers map[string]string
	url     *string
}

func (p *URLPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		url, err := f.Var(*p.url)
		if err != nil {
			return err
		}

		var body io.Reader
		if p.method == http.MethodPost {
			body, err = tap.Reader(f.Object)
			if err != nil {
				return errors.Wrap(err, "url POST")
			}
		}

		req, err := http.NewRequest(p.method, url, body)
		if err != nil {
			return err
		}

		for k, v := range p.headers {
			req.Header.Set(k, v)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		err = stream.Write(&Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       resp.Body,
		})
	}
}
