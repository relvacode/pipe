package console

import (
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
	ptr      reflect.Value
	fallback reflect.Value
	apply    apply
}

func (a *Option) Set(input string) error {
	if !a.ptr.IsValid() || a.apply == nil {
		panic(errors.New("set argument without declared values"))
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
	default:
		panic(errors.Errorf("Cannot set default on %T", a.ptr.Interface()))
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
		switch {
		case x.Kind() == reflect.Interface:
		case x.Type() != a.fallback.Type():
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

func (a *Option) Template() *tap.Template {
	var t tap.Template
	var ptr = &t
	a.init(ptr, func(s string) error {
		*ptr = tap.Template(s)
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
func (a *Option) Expression() *Expression {
	var n Expression
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
