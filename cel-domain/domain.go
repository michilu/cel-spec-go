package cel

import (
	"fmt"
	"reflect"

	"github.com/google/cel-go/checker"
	"github.com/google/cel-go/common/packages"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/interpreter"
	"github.com/google/cel-go/interpreter/functions"
	"github.com/google/cel-go/parser"
	"github.com/google/cel-spec/proto/checked/v1/checked"

	"github.com/michilu/boilerplate/v/errs"
)

type (
	Domain interface {
		Eval(bindings map[string]interface{}) (interface{}, error)
	}

	domain struct {
		interpretable interpreter.Interpretable
		resultType    interface{}
	}
)

func NewDomain(text string, decls []*checked.Decl, returnType interface{}) (Domain, error) {
	const op = "cel.NewDomain"

	// Parse the expression and returns the accumulated errors.
	p, err := parser.ParseText(text)
	if len(err.GetErrors()) != 0 {
		return nil, &errs.Error{Op: op, Message: err.ToDisplayString()}
	}

	// Check the expression matches expectations given the declarations for
	// the identifiers a, b, c where the identifiers are scoped to the default
	// package (empty string):
	typeProvider := types.NewProvider()
	env := checker.NewStandardEnv(packages.DefaultPackage, typeProvider, err)
	env.Add(decls...)
	ch := checker.Check(p, env)
	if len(err.GetErrors()) != 0 {
		return nil, &errs.Error{Op: op, Message: err.ToDisplayString()}
	}

	// Interpret the checked expression using the standard overloads.
	i := interpreter.NewStandardIntepreter(packages.DefaultPackage, typeProvider)
	eval := i.NewInterpretable(interpreter.NewCheckedProgram(ch))

	return &domain{eval, returnType}, nil
}

func NewProgram(text string, decls []*checked.Decl) (interpreter.Program, interpreter.MutableEvalState, error) {
	const op = "cel.NewProgram"

	// Parse the expression and returns the accumulated errors.
	p, err := parser.ParseText(text)
	if len(err.GetErrors()) != 0 {
		return nil, nil, &errs.Error{Op: op, Message: err.ToDisplayString()}
	}

	// Check the expression matches expectations given the declarations for
	// the identifiers a, b, c where the identifiers are scoped to the default
	// package (empty string):
	typeProvider := types.NewProvider()
	env := checker.NewStandardEnv(packages.DefaultPackage, typeProvider, err)
	env.Add(decls...)
	ch := checker.Check(p, env)
	if len(err.GetErrors()) != 0 {
		return nil, nil, &errs.Error{Op: op, Message: err.ToDisplayString()}
	}

	// github.com/google/cel-go/interpreter#NewStandardIntepreter
	d := interpreter.NewDispatcher()
	er := d.Add(functions.StandardOverloads()...)
	if er != nil {
		return nil, nil, &errs.Error{Op: op, Err: er}
	}

	c := interpreter.NewCheckedProgram(ch)

	// github.com/google/cel-go/interpreter.exprInterpreter.NewInterpretable
	e := interpreter.NewEvalState(c.MaxInstructionId() + 1)
	c.Init(d, e)

	return c, e, nil
}

func (d *domain) Eval(m map[string]interface{}) (interface{}, error) {
	const op = "cel.domain.Eval"
	v, _ := d.interpretable.Eval(interpreter.NewActivation(m))
	t := d.resultType
	n, err := v.ConvertToNative(reflect.TypeOf(t))
	if err != nil {
		return nil, &errs.Error{Op: op, Err: err, Message: fmt.Sprintf("execution result types do not match. needs %T. got %#v of %T.", t, n, n)}
	}
	return n, nil
}
