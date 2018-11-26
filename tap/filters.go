package tap

import (
	"bytes"
	"encoding/json"
	"github.com/flosch/pongo2"
)

func init() {
	pongo2.RegisterFilter("mktemp", TempFileFilter)
	pongo2.RegisterFilter("json", JSONFilter)
}

func TempFileFilter(in *pongo2.Value, _ *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	file, err := MkTemp(in.Interface())
	if err != nil {
		return nil, &pongo2.Error{
			OrigError: err,
		}
	}
	return pongo2.AsValue(file), nil
}

func JSONFilter(in *pongo2.Value, _ *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(in.Interface())
	if err != nil {
		return nil, &pongo2.Error{
			OrigError: err,
		}
	}

	return pongo2.AsValue(buf.String()), nil
}
