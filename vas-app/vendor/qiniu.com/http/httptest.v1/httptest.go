package httptest

import (
	"github.com/qiniu/http/httptest.v1"
	"github.com/qiniu/http/httptest.v1/exec"

	_ "qiniu.com/http/httptest.v1/exec/plugin"
)

// ---------------------------------------------------------------------------

type Context struct {
	*httptest.Context
	Ectx *exec.Context
}

func New(t httptest.TestingT) Context {

	ctx := httptest.New(t)
	ectx := exec.New()
	return Context{ctx, ectx}
}

func (p Context) Exec(code string) Context {

	p.Context.Exec(p.Ectx, code)
	return p
}

// ---------------------------------------------------------------------------

