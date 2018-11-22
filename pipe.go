package pipe

import (
	"context"
)

type Pipe interface {
	Go(context.Context, Stream) error
}
