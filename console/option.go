package console

import (
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe/tap"
	"github.com/sirupsen/logrus"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// oType wraps logic around a value (usually a pointer) to defer parsing of that value from an input string.
type oType struct {
	// Name of the type as displayed to a human
	Name string
	// Parse sets the value from an input string
	Parse func(input string) error
	// SetDefault sets the value using a native Go value
	SetDefault func(value reflect.Value)
}

// Options convert a string value provided by the user to pointer value described when a pipe is constructed.
type Option struct {
	name         string
	optionType   *oType
	defaultValue *reflect.Value
}

func (o *Option) Set(input string) error {
	if input == "" {
		if o.defaultValue == nil {
			return errors.Errorf("option %v: missing required argument (%s not provided)", o.name, o.optionType.Name)
		}
		return nil
	}
	return o.optionType.Parse(input)
}

func (o *Option) Usage() string {
	var s strings.Builder
	if o.defaultValue != nil {
		s.WriteString("[")
	}
	s.WriteString("<")
	_, _ = fmt.Fprint(&s, o.optionType.Name)
	if o.defaultValue != nil && o.defaultValue.IsValid() {
		_, _ = fmt.Fprintf(&s, "=%q", fmt.Sprint(o.defaultValue.Interface()))
	}
	s.WriteString(">")
	if o.defaultValue != nil {
		s.WriteString("]")
	}
	return s.String()
}

func (o *Option) assign(t oType) {
	o.optionType = &t
	if o.defaultValue != nil {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("invalid default (%T %v) caused panic on option %s", o.defaultValue.Interface(), o.defaultValue.Interface(), o.name)
				panic(r)
			}
		}()

		o.optionType.SetDefault(*o.defaultValue)
	}
}

// Default sets a default value on this option if the input value is empty.
// This is usually the non-pointer form of the option's pointer value or a map for a map type.
func (o *Option) Default(value interface{}) *Option {
	v := reflect.ValueOf(value)
	o.defaultValue = &v
	return o
}

func (o *Option) String() *string {
	var value string
	var ptr = &value

	o.assign(oType{
		Name: "string",
		Parse: func(input string) error {
			*ptr = input
			return nil
		},
		SetDefault: func(value reflect.Value) {
			*ptr = value.String()
		},
	})
	return ptr
}

func (o *Option) Bool() *bool {
	var value bool
	var ptr = &value

	o.assign(oType{
		Name: "bool",
		Parse: func(input string) error {
			b, err := strconv.ParseBool(input)
			if err != nil {
				return err
			}
			*ptr = b
			return nil
		},
		SetDefault: func(value reflect.Value) {
			*ptr = value.Bool()
		},
	})
	return ptr
}

func (o *Option) Template() *tap.Template {
	var t tap.Template
	var ptr = &t

	o.assign(oType{
		Name: "template",
		Parse: func(input string) error {
			*ptr = tap.Template(input)
			return nil
		},
		SetDefault: func(value reflect.Value) {
			*ptr = tap.Template(value.String())
		},
	})
	return ptr
}

// Int parses as an 64 bit integer
func (o *Option) Int() *int64 {
	var value int64
	var ptr = &value

	o.assign(oType{
		Name: "int",
		Parse: func(input string) error {
			i, err := strconv.ParseInt(input, 10, 64)
			if err != nil {
				return err
			}
			*ptr = i
			return nil
		},
		SetDefault: func(value reflect.Value) {
			*ptr = value.Int()
		},
	})
	return ptr
}

// Expression parses an expr expression.
// Use with a default value of `nil`
func (o *Option) Expression() *Expression {
	var n Expression
	var ptr = &n

	o.assign(oType{
		Name: "expression",
		Parse: func(input string) error {
			e, err := expr.Parse(input)
			if err != nil {
				return err
			}
			*ptr = e
			return nil
		},
		SetDefault: func(value reflect.Value) {
			*ptr = value.Interface().(Expression)
		},
	})
	return ptr
}

// Duration parses a duration
func (o *Option) Duration() *time.Duration {
	var d time.Duration
	var ptr = &d

	o.assign(oType{
		Name: "duration",
		Parse: func(input string) error {
			pd, err := time.ParseDuration(input)
			if err != nil {
				return err
			}
			*ptr = pd
			return nil
		},
		SetDefault: func(value reflect.Value) {
			*ptr = value.Interface().(time.Duration)
		},
	})
	return ptr
}

func (o *Option) Map() map[string]string {
	var ptr = make(map[string]string)

	o.assign(oType{
		Name: "key:value",
		Parse: func(input string) error {
			parts := strings.Split(input, ":")
			if len(parts) < 2 {
				return errors.Errorf("expected key:value in %q", input)
			}
			ptr[parts[0]] = strings.Join(parts[1:], ":")
			return nil
		},
		SetDefault: func(value reflect.Value) {
			if !value.IsValid() || value.IsNil() {
				return
			}
			for k, v := range value.Interface().(map[string]string) {
				ptr[k] = v
			}
		},
	})
	return ptr
}
