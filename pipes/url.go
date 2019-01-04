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
	for i := range methods {
		method := methods[i]
		pipe.Define(pipe.Pkg{
			Name: pipe.Family("url", strings.ToLower(method)),
			Constructor: func(console *console.Command) pipe.Pipe {
				return &URLPipe{
					method:  method,
					headers: console.Option("header").Default(nil).Map(),
					body:    console.Option("body").Default(false).Bool(),
					url:     console.Arg(0).Template(),
				}
			},
		})
	}
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
	body    *bool
	url     *tap.Template
}

func (p *URLPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		url, err := p.url.Render(f.Context())
		if err != nil {
			return err
		}

		var body io.Reader
		if *p.body {
			body, err = tap.Reader(f.Object)
			if err != nil {
				return errors.Wrap(err, "cannot use this input as the body of the request")
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

		err = stream.Write(nil, &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       resp.Body,
		})
		if err != nil {
			return err
		}
	}
}
