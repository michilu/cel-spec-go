package example

import (
	"fmt"
	"strings"

	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-spec/proto/checked/v1/checked"
	"github.com/michilu/boilerplate/v/errs"

	cel "github.com/michilu/cel-spec-go/cel-domain"
)

const (
	paypalmeCel = `
"https://www.paypal.me/" + username + "/" + amount + currency
`
	paypalmeReturnType = "string"
)

var (
	paypalmeParams = []*checked.Decl{
		decls.NewIdent("username", decls.String, nil),
		decls.NewIdent("amount", decls.String, nil),
		decls.NewIdent("currency", decls.String, nil),
	}

	domain cel.Domain
)

func init() {
	d, err := cel.NewDomain(paypalmeCel, paypalmeParams, paypalmeReturnType)
	if err != nil {
		panic(err)
	}
	domain = d
}

func example(u, a, c string) (string, error) {
	const op = "example.example"
	v, err := domain.Eval(map[string]interface{}{
		"u": u,
		"a": a,
		"c": c,
	})
	if err != nil {
		return "", &errs.Error{Op: op, Err: err}
	}
	s, ok := v.(string)
	if !ok {
		return "", &errs.Error{Op: op, Message: fmt.Sprintf("execution result types do not match. needs %T. got %#v of %T.", "string", v, v)}
	}
	return strings.ToLower(s), nil
}
