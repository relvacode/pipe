package console

import (
	"github.com/antonmedv/expr"
	"github.com/google/shlex"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type apply func(string) error

// Option describes the conversion of an input string into a destination pointer value.
// Arguments are required by default unless a default value is supplied.
type Option struct {
	ptr      reflect.Value
	fallback reflect.Value
	apply    apply
}

func (a *Option) Set(input string) error {
	if !a.ptr.IsValid() || a.apply == nil {
		panic("set argument without declared values")
	}
	if input != "" {
		return a.apply(input)
	}

	if !a.fallback.IsValid() {
		return errors.New("value required")
	}
	switch a.ptr.Kind() {
	case reflect.Ptr:

		a.ptr.Elem().Set(a.fallback)
	case reflect.Map:
		keys := a.fallback.MapKeys()
		for _, k := range keys {
			a.ptr.SetMapIndex(k, a.fallback.MapIndex(k))
		}
	}
	return nil
}

func (a *Option) init(ptr interface{}, f apply) {
	a.ptr = reflect.ValueOf(ptr)
	a.apply = f

	if a.fallback.IsValid() {
		var x = a.ptr
		if x.Kind() == reflect.Ptr {
			x = x.Elem()
		}
		if x.Type() != a.fallback.Type() {
			panic(errors.Errorf("invalid default type %s for argument type %s", a.fallback.Type(), x.Type()))
		}
	}
}

func (a *Option) Default(v interface{}) *Option {
	a.fallback = reflect.ValueOf(v)
	return a
}

// String is a optional string
func (a *Option) String() *string {
	var value string
	var ptr = &value
	a.init(ptr, func(s string) error {
		*ptr = s
		return nil
	})
	return ptr
}

// Int parses as an 64 bit integer
func (a *Option) Int() *int64 {
	var value int64
	var ptr = &value
	a.init(ptr, func(s string) error {
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
func (a *Option) Expression() *expr.Node {
	var n expr.Node
	var ptr = &n
	a.init(ptr, func(s string) error {
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
func (a *Option) Duration() *time.Duration {
	var d time.Duration
	var ptr = &d
	a.init(ptr, func(s string) error {
		pd, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		*ptr = pd
		return nil
	})
	return ptr
}

func (a *Option) Map() map[string]string {
	var ptr = make(map[string]string)
	a.init(ptr, func(s string) error {
		parts := strings.Split(s, ":")
		if len(parts) < 2 {
			return errors.Errorf("expected key:value in %q", s)
		}
		ptr[parts[0]] = strings.Join(parts[1:], ":")
		return nil
	})
	return ptr
}

// Shell returns all arguments split using a shell-like parser
func (a *Option) Shell() *[]string {
	var args = make([]string, 0)
	var ptr = &args
	a.init(ptr, func(s string) error {
		parsed, err := shlex.Split(s)
		if err != nil {
			return err
		}
		for _, p := range parsed {
			*ptr = append(*ptr, p)
		}
		return nil
	})
	return ptr
}
