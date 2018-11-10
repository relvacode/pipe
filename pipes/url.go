package pipes

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/valve"
	"io"
	"net/http"
)

func init() {
	pipe.Pipes.Define(pipe.ModuleDefinition{
		Name: "url",
		Constructor: func(valve *valve.Control) pipe.Pipe {
			return &URLPipe{
				url: valve.All().String(),
			}
		},
	})
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
	url *string
}

func (p URLPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		url, err := f.Var(*p.url)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
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
