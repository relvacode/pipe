package iterate

import (
	"context"
	"fmt"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"github.com/sirupsen/logrus"
	"os"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "ask",
		Constructor: func(command *console.Command) pipe.Pipe {
			return AskPipe{
				question: command.Any().Template(),
			}
		},
	})
}

type AskPipe struct {
	question *tap.Template
}

func (p AskPipe) Go(ctx context.Context, stream pipe.Stream) error {
stream:
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		q, err := (*p.question).Render(f.Context())
		if err != nil {
			return err
		}

	ask:
		for {
			_, err = fmt.Fprint(os.Stderr, q, " [y/n]: ")
			if err != nil {
				return err
			}

			var reply string
			_, err = fmt.Scanln(&reply)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(os.Stderr)
			if err != nil {
				return err
			}

			switch reply {
			case "Y", "y", "YES", "yes":
				err = stream.Write(nil, f.Object)
				if err != nil {
					return err
				}
				break ask
			case "N", "n", "NO", "no":
				continue stream
			default:
				logrus.Error("please enter yes or no")
			}
		}

	}
}
