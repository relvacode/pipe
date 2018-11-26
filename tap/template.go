package tap

import (
	"github.com/flosch/pongo2"
)

var engine = pongo2.NewSet("pipe", pongo2.MustNewLocalFileSystemLoader(""))

func init() {
	pongo2.SetAutoescape(false)
}

type Template string

// Render the template.
func (t Template) Render(ctx pongo2.Context) (string, error) {
	ts, err := engine.FromString(string(t))
	if err != nil {
		return "", err
	}
	return ts.Execute(ctx)
}

type TemplateSet []string

func (s TemplateSet) Render(ctx pongo2.Context) ([]string, error) {
	var err error
	var vars = make([]string, len(s))

	for i := 0; i < len(s); i++ {
		vars[i], err = Template(s[i]).Render(ctx)
		if err != nil {
			return nil, err
		}
	}
	return vars, nil
}
