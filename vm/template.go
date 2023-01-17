package vm

import (
	"bytes"
	htmltemplate "html/template"
	"text/template"

	"github.com/gentee/gentee/core"
)

func ExecTmpl(rt *Runtime, name string, tmpl string, input *core.Obj) (string, error) {
	cmp, err := template.New(name).Parse(tmpl)
	if err != nil {
		return "", err
	}
	w := new(bytes.Buffer)
	err = cmp.Execute(w, input.Data)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}

func ExecTmplHTML(rt *Runtime, name string, tmpl string, input *core.Obj) (string, error) {
	cmp, err := htmltemplate.New(name).Parse(tmpl)
	if err != nil {
		return "", err
	}
	w := new(bytes.Buffer)
	err = cmp.Execute(w, input.Data)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}
