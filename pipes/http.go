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
		Name: "http",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &HTTPServerPipe{
				address: console.Arg(0).Default("127.0.0.1:8080").String(),
			}
		},
	})
}

type Request struct {
	Method     string
	RemoteAddr string
	Host       string
	URL        string
	Path       string
	Query      url.Values
	Headers    http.Header

	body *bytes.Buffer
}

func (r *Request) String() string {
	return r.body.String()
}

func (r *Request) Read(b []byte) (int, error) {
	return r.body.Read(b)
}

type HTTPServerPipe struct {
	address *string
}

func (p *HTTPServerPipe) handle(stream pipe.Stream) (chan error, http.HandlerFunc) {
	var err = make(chan error, 1)
	return err, func(rw http.ResponseWriter, r *http.Request) {
		req := &Request{
			Method:     r.Method,
			RemoteAddr: r.RemoteAddr,
			Host:       r.Host,
			URL:        r.URL.String(),
			Path:       r.URL.Path,
			Query:      r.URL.Query(),
			Headers:    r.Header,
			body:       new(bytes.Buffer),
		}
		io.Copy(req.body, r.Body)
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
