package pipes

import (
	"github.com/nats-io/go-nats"
	"net/url"
	"context"
	"github.com/relvacode/pipe"
	"strings"
	"bytes"
	"github.com/relvacode/pipe/console"
	"github.com/sirupsen/logrus"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name:        "nats",
		Description: "Publish and subscribe messages to a NATS event channel",
		Family: []pipe.Pkg{
			{
				Name: "subscribe",
				Constructor: func(args *console.Command) pipe.Pipe {
					return &NatsReceiverPipe{
						NatsClient{
							url: args.Input().String(),
						},
					}
				},
			},
			{
				Name: "publish",
				Constructor: func(args *console.Command) pipe.Pipe {
					return &NatsSenderPipe{
						NatsClient{
							url: args.Input().String(),
						},
					}
				},
			},
		},
	})
}

type NatsClient struct {
	url *string
}

func (p NatsClient) Connect() (string, *nats.Conn, error) {
	u, err := url.Parse(*p.url)
	if err != nil {
		return "", nil, err
	}
	cu := &url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
	}
	logrus.Debugf("nats connect %q", cu.String())
	c, err := nats.Connect(cu.String())
	if err != nil {
		return "", nil, err
	}

	return strings.TrimPrefix(u.Path, "/"), c, nil
}

type NatsReceiverPipe struct {
	NatsClient
}

func (p *NatsReceiverPipe) Go(ctx context.Context, stream pipe.Stream) error {
	q, c, err := p.Connect()
	if err != nil {
		return err
	}

	defer c.Close()
	var msg = make(chan *nats.Msg)

	s, err := c.ChanSubscribe(q, msg)
	if err != nil {
		return err
	}

	defer s.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case m := <-msg:
			copied := make([]byte, len(m.Data))
			copy(copied, m.Data)

			err = stream.Write(bytes.NewReader(copied))
			if err != nil {
				return err
			}
		}
	}
}

type NatsSenderPipe struct {
	NatsClient
}

func (p *NatsSenderPipe) Go(ctx context.Context, stream pipe.Stream) error {
	q, c, err := p.Connect()
	if err != nil {
		return err
	}

	defer c.Close()
	var buf bytes.Buffer

	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		_, err = f.WriteTo(&buf)
		if err != nil {
			return err
		}

		err = c.Publish(q, buf.Bytes())
		if err != nil {
			return err
		}
		buf.Reset()
	}
}
