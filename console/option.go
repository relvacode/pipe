package console

import (
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe/tap"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type apply func(string) error

// Options convert a string value provided by the user to pointer value described when a pipe is constructed.
type Option struct {
	ptr   reflect.Value
	def   *string
	apply apply
	usage string
	set   bool
}

func (o *Option) Set(input string) error {
	if !o.ptr.IsValid() || o.apply == nil {
		panic(errors.New("set argument without declared values"))
	}
	if input == "" && o.def == nil {
		return errors.New("required argument")
	}
	if input == "" && o.def != nil {
		input = *o.def
	}
	o.set = true
	return o.apply(input)
}

func (o *Option) Usage() string {
	var s strings.Builder
	if o.def != nil {
		s.WriteString("[")
	}
	s.WriteString("<")
	s.WriteString(o.usage)
	if o.def != nil {
		fmt.Fprintf(&s, "=%q", *o.def)
	}
	s.WriteString(">")
	if o.def != nil {
		s.WriteString("]")
	}
	return s.String()
}

func (o *Option) init(ptr interface{}, t string, f apply) {
	o.ptr = reflect.ValueOf(ptr)
	o.usage = t
	o.apply = f
}

func (o *Option) Default(s string) *Option {
	o.def = &s
	return o
}

// String is a optional string
func (o *Option) String() *string {
	var value string
	var ptr = &value
	o.init(ptr, "string", func(s string) error {
		*ptr = s
		return nil
	})
	return ptr
}

func (o *Option) Template() *tap.Template {
	var t tap.Template
	var ptr = &t
	o.init(ptr, "template", func(s string) error {
		*ptr = tap.Template(s)
		return nil
	})
	return ptr
}

// Int parses as an 64 bit integer
func (o *Option) Int() *int64 {
	var value int64
	var ptr = &value
	o.init(ptr, "int", func(s string) error {
		var err error
		*ptr, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		return nil
	})
	return ptr
}

// Expression parses an expr expression
func (o *Option) Expression() *Expression {
	var n Expression
	var ptr = &n
	o.init(ptr, "expression", func(s string) error {
		pn, err := expr.Parse(s)
		if err != nil {
			return err
		}
		*ptr = pn
		return nil
	})
	return ptr
}

// Duration parses a duration
func (o *Option) Duration() *time.Duration {
	var d time.Duration
	var ptr = &d
	o.init(ptr, "duration", func(s string) error {
		pd, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		*ptr = pd
		return nil
	})
	return ptr
}

func (o *Option) Map() map[string]string {
	var ptr = make(map[string]string)
	o.init(ptr, "key:value", func(s string) error {
		parts := strings.Split(s, ":")
		if len(parts) < 2 {
			return errors.Errorf("expected key:value in %q", s)
		}
		ptr[parts[0]] = strings.Join(parts[1:], ":")
		return nil
	})
	return ptr
}
