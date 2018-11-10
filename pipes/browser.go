package pipes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/valve"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func init() {
	pipe.Pipes.Define(pipe.ModuleDefinition{
		Name: "http.events",
		Constructor: func(valve *valve.Control) pipe.Pipe {
			return &BrowserPipe{
				addr: valve.All().DefaultString("127.0.0.1:3003"),
				data: make(chan string),
			}
		},
	})
}

var (
	index = []byte(`
<!DOCTYPE html>
<html>
<head>
	<style>
html {
    background-color: #1b1b1b;
    color: #f9f9f9;
    line-height: 1em;
}
div#content {
    display: flex;
    flex-direction: column;
}
	</style>
</head>
<body>
    <div id="content"></div>
    <script type="text/javascript">
        var source = new EventSource("http://localhost:3000/events");
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
`)
)

// BrowserPipe streams each input item to your browser
type BrowserPipe struct {
	addr *string
	data chan string
}

func (p *BrowserPipe) indexHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	io.Copy(rw, bytes.NewReader(index))
}

func (p *BrowserPipe) eventsHandler(rw http.ResponseWriter, r *http.Request) {
	// Listen to the closing of the http connection via the CloseNotifier
	notify := rw.(http.CloseNotifier).CloseNotify()
	f := rw.(http.Flusher)

	// Set the headers related to event streaming.
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Transfer-Encoding", "chunked")
	rw.WriteHeader(http.StatusOK)

	e := json.NewEncoder(rw)
	for {
		select {
		case <-notify:
			return
		case message, ok := <-p.data:
			if !ok {
				return
			}
			fmt.Fprint(rw, "data: ")
			err := e.Encode(message)
			if err != nil {
				logrus.Error(err)
				return
			}
			fmt.Fprint(rw, "\n\n")
			f.Flush()
		}
	}
}

func (p *BrowserPipe) start(ctx context.Context) chan error {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", p.eventsHandler)
	mux.HandleFunc("/", p.indexHandler)

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
		server.Shutdown(ctx)
	}()

	return errors
}

func (p *BrowserPipe) Go(ctx context.Context, stream pipe.Stream) error {
	ctx, cancel := context.WithCancel(ctx)
	defer close(p.data)
	defer cancel()

	errors := p.start(ctx)

	var buf = new(bytes.Buffer)
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}
		_, err = f.WriteTo(buf)
		if err != nil {
			return err
		}

		select {
		case p.data <- buf.String():
		case err := <-errors:
			return err
		}

		buf.Reset()
	}
}
