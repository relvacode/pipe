package pipe

import (
	"fmt"
	"github.com/flosch/pongo2"
	"io"
)

func NewTag(t string) *Tag {
	if t == "" {
		return nil
	}
	tag := Tag(t)
	return &tag
}

type Tag string

func (t *Tag) String() string {
	if t == nil {
		return "<none>"
	}
	return string(*t)
}

type Stack map[string]interface{}

func NewDataFrame(x interface{}, tag *Tag) *DataFrame {
	return &DataFrame{
		Tag:    tag,
		Object: x,
		Stack:  make(Stack),
	}
}

type DataFrame struct {
	Tag    *Tag
	Object interface{}
	Index  uint64
	Stack  Stack

	context pongo2.Context // cached context
}

func (f *DataFrame) String() string {
	return fmt.Sprintf("DataFrame(%s: %T: %d refs)", f.Tag, f.Object, len(f.Stack))
}

// AppendStack creates a copy of this DataFrame with additional stack context
func (f *DataFrame) AppendStack(s Stack) *DataFrame {
	nf := f.Copy(f.Object, f.Tag)
	if s != nil {
		for k, v := range s {
			nf.Stack[k] = v
		}
	}
	return nf
}

// Copy copies this data frame
func (f *DataFrame) Copy(x interface{}, tag *Tag) *DataFrame {
	l := len(f.Stack)
	if f.Tag != nil {
		l++
	}
	n := &DataFrame{
		Tag:    tag,
		Object: x,
		Stack:  make(Stack, len(f.Stack)),
	}
	for k, v := range f.Stack {
		n.Stack[k] = v
	}
	if f.Tag != nil {
		n.Stack[string(*f.Tag)] = f.Object
	}
	return n
}

var _ io.WriterTo = (*DataFrame)(nil)

// Writes the contents of this frame object to the given writer
func (f *DataFrame) WriteTo(w io.Writer) (int64, error) {
	switch o := f.Object.(type) {
	case io.Reader:
		return io.Copy(w, o)
	case io.WriterTo:
		return o.WriteTo(w)
	default:
		i, err := fmt.Fprint(w, o)
		return int64(i), err
	}
}

func (f *DataFrame) Context() pongo2.Context {
	if f.context != nil {
		return f.context
	}
	f.context = make(pongo2.Context, len(f.Stack)+1)
	for k, v := range f.Stack {
		f.context[k] = v
	}
	if f.Tag != nil {
		f.context[string(*f.Tag)] = f.Object
	}
	f.context["this"] = f.Object
	f.context["_index"] = f.Index
	return f.context
}
