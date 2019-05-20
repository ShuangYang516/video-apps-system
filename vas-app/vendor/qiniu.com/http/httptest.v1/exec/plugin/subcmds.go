package plugin

import (
	"encoding/base64"
	"encoding/binary"
	"net/http"
	"reflect"
	"strconv"
	"syscall"

	"qiniu.com/auth/authstub.v1"

	"github.com/qiniu/http/httptest.v1"
	. "github.com/qiniu/http/httptest.v1/exec"
	. "qiniu.com/auth/proto.v1"
)

// ---------------------------------------------------------------------------

type subContext struct {
	ctx IContext
}

func init() {

	ExternalSub = new(subContext)
}

func (p *subContext) FindCmd(ctx IContext, cmd string) reflect.Value {

	p.ctx = ctx
	v := reflect.ValueOf(p)
	return v.MethodByName("Eval_" + cmd)
}

// ---------------------------------------------------------------------------

type authstubArgs struct {
	Uid     uint   `flag:"uid"`
	Utype   uint   `flag:"utype"`
	Sudoer  uint   `flag:"suid"`
	UtypeSu uint   `flag:"sut"`
	App     string `arg:"app,opt"`
}

type authstubTransportComposer struct {
	args *authstubArgs
	ctx  *httptest.Context
}

func toAppId(app string) (appId uint64, err error) {

	b, err := base64.URLEncoding.DecodeString(app)
	if err != nil {
		return
	}
	if len(b) != 12 {
		return 0, syscall.EINVAL
	}
	return binary.LittleEndian.Uint64(b[4:]), nil
}

func (p *authstubTransportComposer) Compose(base http.RoundTripper) http.RoundTripper {

	var appId uint64
	var err error

	if p.args.App != "" {
		appId, err = strconv.ParseUint(p.args.App, 10, 64)
		if err != nil {
			appId, err = toAppId(p.args.App)
			if err != nil {
				p.ctx.Fatal("Parse arg `app` failed:", err)
			}
		}
	}
	ui := &SudoerInfo{
		UserInfo: UserInfo{
			Uid: uint32(p.args.Uid),
			Utype: uint32(p.args.Utype),
			Appid: appId,
		},
		Sudoer: uint32(p.args.Sudoer),
		UtypeSu: uint32(p.args.UtypeSu),
	}
	return authstub.NewTransport(ui, base)
}

func (p *subContext) Eval_authstub(
	ctx *httptest.Context, args *authstubArgs) (httptest.TransportComposer, error) {

	return &authstubTransportComposer{args, ctx}, nil
}

// ---------------------------------------------------------------------------

