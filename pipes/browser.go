package pipes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"github.com/sirupsen/logrus"
	"html/template"
	"io"
	"net/http"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "browser",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &BrowserPipe{
				addr: console.Arg(0).Default("127.0.0.1:3003").String(),
				fg:   console.Option("fg").Default("#f9f9f9").String(),
				bg:   console.Option("bg").Default("#1b1b1b").String(),
				data: make(chan *BufSend),
			}
		},
	})
}

var (
	index = `
<!DOCTYPE html>
<html>
<head>
	<style>
html {
    background-color: {{.bg}};
    color: {{.fg}};
    line-height: 1em;
}
div#content {
    display: flex;
    flex-direction: column;
}
code {
	white-space: pre-wrap;
}
	</style>
</head>
<body>
    <div id="content"></div>
    <script type="text/javascript">
        var source = new EventSource("http://{{.addr}}/events");
        source.onmessage = function(event) {
            var content = document.getElementById('content');
			var obj = JSON.parse(event.data)
			var el = document.createElement("code");
			el.innerHTML = obj;
            content.appendChild(el);
        };
    </script>
</body>
</html>
`
)

type BufSend struct {
	Buf   *bytes.Buffer
	Reply chan *bytes.Buffer
}

// BrowserPipe streams each input item to your browser
type BrowserPipe struct {
	fg, bg *string
	addr   *string
	data   chan *BufSend
}

func (p *BrowserPipe) indexHandler() http.HandlerFunc {
	t, err := template.New("src").Parse(index)
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer
	err = t.Execute(&b, map[string]interface{}{
		"addr": *p.addr,
		"fg":   *p.fg,
		"bg":   *p.bg,
	})

	if err != nil {
		panic(err)
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/html")
		rw.WriteHeader(http.StatusOK)
		_, _ = io.Copy(rw, bytes.NewReader(b.Bytes()))
	}
}

func (p *BrowserPipe) eventsHandler(rw http.ResponseWriter, r *http.Request) {
	// Listen to the closing Of the http connection via the CloseNotifier
	f := rw.(http.Flusher)

	// Set the headers related to event streaming.
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Command", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Transfer-Encoding", "chunked")
	rw.WriteHeader(http.StatusOK)

	e := json.NewEncoder(rw)
	for {
		select {
		case <-r.Context().Done():
			return
		case b, ok := <-p.data:
			if !ok {
				return
			}

			_, _ = fmt.Fprint(rw, "data: ")
			err := e.Encode(b.Buf.String())
			b.Reply <- b.Buf

			if err != nil {
				logrus.Error(err)
				return
			}
			_, _ = fmt.Fprint(rw, "\n\n")
			f.Flush()
		}
	}
}

func (p *BrowserPipe) start(ctx context.Context) chan error {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", p.eventsHandler)
	mux.HandleFunc("/", p.indexHandler())

	server := &http.Server{
		Handler: mux,
		Addr:    *p.addr,
	}

	errors := make(chan error, 1)
	go func() {
		errors <- server.ListenAndServe()
	}()

	go func() {
		<-ctx.Done()
		tap.LogError(server.Shutdown(ctx))
	}()

	return errors
}

func (p *BrowserPipe) Go(ctx context.Context, stream pipe.Stream) error {
	ctx, cancel := context.WithCancel(ctx)
	defer close(p.data)
	defer cancel()

	errors := p.start(ctx)
	logrus.Infof("Open %q in your browser", *p.addr)

	b := &BufSend{
		Buf:   new(bytes.Buffer),
		Reply: make(chan *bytes.Buffer),
	}
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}
		_, err = f.WriteTo(b.Buf)
		if err != nil {
			return err
		}

		select {
		case p.data <- b:
			b.Buf = <-b.Reply
			b.Buf.Reset()
		case err := <-errors:
			return err
		}
	}
}
