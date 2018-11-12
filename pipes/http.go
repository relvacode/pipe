package pipes

import (
	"bytes"
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"io"
	"net/http"
	"net/url"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "http.request",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &HTTPServerPipe{
				address: console.Default("127.0.0.1:3002").String(),
			}
		},
	})
}

type Request struct {
	Method  string
	URL     string
	Path    string
	Query   url.Values
	Headers http.Header
	Body    *bytes.Buffer
}

func (r *Request) Read(b []byte) (int, error) {
	return r.Body.Read(b)
}

type HTTPServerPipe struct {
	address *string
}

func (p *HTTPServerPipe) handle(stream pipe.Stream) (chan error, http.HandlerFunc) {
	var err = make(chan error, 1)
	return err, func(rw http.ResponseWriter, r *http.Request) {
		req := &Request{
			Method:  r.Method,
			URL:     r.URL.String(),
			Path:    r.URL.Path,
			Query:   r.URL.Query(),
			Headers: r.Header,
			Body:    new(bytes.Buffer),
		}
		io.Copy(req.Body, r.Body)
		r.Body.Close()
		e := stream.Write(req)
		if e != nil {
			err <- e
		}
	}
}

func (p *HTTPServerPipe) Go(ctx context.Context, stream pipe.Stream) error {
	cancel, handler := p.handle(stream)
	errors := make(chan error, 1)
	server := &http.Server{
		Addr:    *p.address,
		Handler: handler,
	}
	go func() {
		errors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		server.Shutdown(nil)
		return nil
	case err := <-cancel:
		server.Shutdown(ctx)
		return err
	case err := <-errors:
		return err
	}
}
