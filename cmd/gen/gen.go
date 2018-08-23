package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/interpreter"
	"github.com/google/cel-spec/proto/checked/v1/checked"
	"github.com/lkesteloot/astutil"
	"github.com/michilu/boilerplate/v/errs"
	"github.com/michilu/boilerplate/v/log"

	cel "github.com/michilu/cel-spec-go/cel-domain"
)

func gen(s string) {
	const op = "cmd.gen.gen"
	var err error

	if s == "" {
		s, err = os.Getwd()
		if err != nil {
			log.Logger().Fatal().
				Str("op", op).
				Err(&errs.Error{Op: op, Err: err}).
				Msg("error")
		}
	}

	log.Logger().Debug().
		Str("op", op).
		Msg(s)

	ps := make([]string, 0)

	err = filepath.Walk(s, func(p string, i os.FileInfo, err error) error {
		const op = "cmd.gen.gen:filepath.Walk"
		if err != nil {
			return &errs.Error{Op: op, Err: err}
		}

		if i.IsDir() {
			return nil
		}
		switch {
		// must be has '.cel' file extension, or includes '.cel.' in the filename
		case strings.HasSuffix(p, ".cel"),
			strings.Contains(path.Base(p), ".cel."):
			// pass
		default:
			return nil
		}

		ps = append(ps, p)
		return nil
	})
	if err != nil {
		log.Logger().Fatal().
			Str("op", op).
			Err(&errs.Error{Op: op, Err: err}).
			Msg("error")
	}

	m := map[string][]string{}
	for _, v := range ps {
		k := path.Dir(v)
		if m[k] == nil {
			m[k] = []string{v}
			continue
		}
		m[k] = append(m[k], v)
	}

	for k, v := range m {
		f, err := getFile(k)
		if err != nil {
			log.Logger().Error().
				Str("op", op).
				Err(&errs.Error{Op: op, Err: err}).
				Msg("error")
		}
		var (
			pkgSign = fmt.Sprintf("package %s", f.Name.Name)
		)
		for _, i := range v {
			var (
				fullpath = i
				filename = filepath.Base(fullpath)
				basename = strings.SplitN(filename, ".cel", 2)[0]
				typeName = strings.Title(basename)
				funcName = fmt.Sprintf("%sFunc", typeName)
				output   = filepath.Join(k, fmt.Sprintf("%s_gen.go", basename))
			)
			df := astutil.DuplicateFile(f)
			ok := ast.FilterFile(df, func(s string) bool { return s == typeName })
			if !ok {
				log.Logger().Error().
					Str("op", op).
					Err(&errs.Error{Op: op, Message: fmt.Sprintf("not found a interface definition of the '%s'", typeName)}).
					Msg("error")
				continue
			}

			var (
				ft *ast.FuncType
				c  = make([]*checked.Decl, 0)
				t  string
			)

			ast.Inspect(df, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncType:
					ft = astutil.DuplicateFuncType(x)

					for _, x := range x.Params.List {
						var (
							t *checked.Type
						)
						switch x.Type.(*ast.Ident).Name {
						case "string":
							t = decls.String
						}
						for _, i := range x.Names {
							c = append(c, decls.NewIdent(i.Name, t, nil))
						}
					}

					if x.Results == nil {
						return false
					}
					if x.Results.List == nil {
						return false
					}
					l := len(x.Results.List)
					if l != 1 {
						log.Logger().Fatal().
							Str("op", op).
							Err(&errs.Error{Op: op, Message: fmt.Sprintf("must returns one value, it returns %d value", l)}).
							Msg("error")
						return false
					}
					i, ok := x.Results.List[0].Type.(*ast.Ident)
					if !ok {
						return false
					}
					t = i.Name

					return false
				}
				return true
			})

			log.Logger().Debug().
				Str("op", op).
				Str("return type", t).
				Msg("debug")

			bi, err := ioutil.ReadFile(fullpath) // #nosec
			if err != nil {
				log.Logger().Error().
					Str("op", op).
					Err(&errs.Error{Op: op, Err: err}).
					Msg("error")
				continue
			}

			prog, state, err := cel.NewProgram(string(bi), c)
			if err != nil {
				log.Logger().Error().
					Str("op", op).
					Err(&errs.Error{Op: op, Err: err}).
					Msg("error")
				continue
			}

			e, err := eval(prog, state)
			if err != nil {
				log.Logger().Error().
					Str("op", op).
					Err(&errs.Error{Op: op, Err: err}).
					Msg("error")
				continue
			}

			fd := []ast.Decl{
				&ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{
								&ast.Ident{Name: funcName},
							},
							Type: &ast.Ident{Name: typeName},
							Values: []ast.Expr{
								&ast.FuncLit{
									Type: ft,
									Body: &ast.BlockStmt{
										List: []ast.Stmt{
											&ast.ReturnStmt{
												Results: e,
											},
										},
									},
								},
							},
						},
					},
				},
			}

			fset := token.NewFileSet()
			pf, err := parser.ParseFile(fset, "", pkgSign, 0)
			if err != nil {
				log.Logger().Error().
					Str("op", op).
					Err(&errs.Error{Op: op, Err: err}).
					Msg("error")
				continue
			}
			pf.Decls = append(pf.Decls, fd...)

			var b bytes.Buffer
			err = format.Node(&b, fset, pf)
			if err != nil {
				log.Logger().Error().
					Str("op", op).
					Err(&errs.Error{Op: op, Err: err}).
					Msg("error")
				continue
			}

			err = ioutil.WriteFile(output, b.Bytes(), 0644)
			if err != nil {
				log.Logger().Error().
					Str("op", op).
					Err(&errs.Error{Op: op, Err: err}).
					Msg("error")
				continue
			}

		}
	}

}

func getFile(d string) (*ast.File, error) {
	const op = "cmd.gen.getFile"

	di, err := os.Stat(d)
	if err != nil {
		return nil, &errs.Error{Op: op, Err: err}
	}
	if !di.IsDir() {
		return nil, &errs.Error{Op: op, Message: "must be a path to an exists directory"}
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, d, nil, 0)
	if err != nil {
		return nil, &errs.Error{Op: op, Err: err}
	}
	if len(pkgs) != 1 {
		return nil, &errs.Error{Op: op, Message: fmt.Sprintf("must be one package, %d founds", len(pkgs))}
	}
	for _, pkg := range pkgs {
		f := ast.MergePackageFiles(pkg, ast.FilterFuncDuplicates|ast.FilterImportDuplicates)
		return f, nil
	}
	return nil, &errs.Error{Op: op, Message: "not found package"}
}

func eval(prog interpreter.Program, state interpreter.MutableEvalState) ([]ast.Expr, error) {
	const op = "cmd.gen.eval"

	var (
		e = make([]ast.Expr, 0)
		m = make(map[int64]string)
	)

	// github.com/google/cel-go/interpreter.exprInterpretable.Eval
	stepper := prog.Begin()
	for step, hasNext := stepper.Next(); hasNext; step, hasNext = stepper.Next() {

		log.Logger().Debug().
			Str("op", op).
			Msgf("step: %#v", step)

		switch step.(type) {
		case *interpreter.IdentExpr:
			x := step.(*interpreter.IdentExpr)
			m[x.GetId()] = x.Name
		case *interpreter.CallExpr:
			x := step.(*interpreter.CallExpr)
			ne, err := callExpr(x, state, m)
			if err != nil {
				log.Logger().Error().
					Str("op", op).
					Err(err).
					Msg("error")
				continue
			}
			e = []ast.Expr{ne}

			//case *interpreter.CreateListExpr:
			//	i.evalCreateList(step.(*interpreter.CreateListExpr))
			//case *interpreter.CreateMapExpr:
			//	i.evalCreateMap(step.(*interpreter.CreateMapExpr))
			//case *interpreter.CreateObjectExpr:
			//	i.evalCreateType(step.(*interpreter.CreateObjectExpr))
			//case *interpreter.MovInst:
			//	i.evalMov(step.(*interpreter.MovInst))
			//	// Special instruction for modifying the program cursor
			//case *interpreter.JumpInst:
			//	jmpExpr := step.(*interpreter.JumpInst)
			//	if jmpExpr.OnCondition(i.state) {
			//		if !stepper.JumpCount(jmpExpr.Count) {
			//			// TODO: Error, the jump count should never exceed the
			//			// program length.
			//			panic("jumped too far")
			//		}
			//	}
			//	// Special instructions for modifying the activation stack
			//case *interpreter.PushScopeInst:
			//	pushScope := step.(*interpreter.PushScopeInst)
			//	scopeDecls := pushScope.Declarations
			//	childActivaton := make(map[string]interface{})
			//	for key, declId := range scopeDecls {
			//		childActivaton[key] = func() interface{} {
			//			return i.value(declId)
			//		}
			//	}
			//	currActivation = NewHierarchicalActivation(currActivation, NewActivation(childActivaton))
			//case *interpreter.PopScopeInst:
			//	currActivation = currActivation.Parent()
		default:
			log.Logger().Debug().
				Str("op", op).
				Msgf("step: %#v", step)
		}

		//s = append(s, "")
	}

	log.Logger().Debug().
		Str("op", op).
		Msgf("m: %#v", m)

	return e, nil
}

var b *ast.BinaryExpr
var pb *ast.BinaryExpr

func callExpr(x *interpreter.CallExpr, state interpreter.MutableEvalState, m map[int64]string) (ast.Expr, error) {
	const op = "callExpr"

	switch x.Function {
	case "_+_":
		pb = b
		b = &ast.BinaryExpr{Op: token.ADD}

		for i, a := range x.Args {

			if i > 1 {
				log.Logger().Error().
					Str("op", op).
					Msg("error")
				break
			}

			o, ok := state.Value(a)
			if !ok {
				log.Logger().Error().
					Str("op", op).
					Msgf("not found IdentExpr %#v in %#v", a, state)
				continue
			}
			if o != nil {
				l := &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%#v", o.Value().(string))}
				switch i {
				case 0:
					b.X = l
				case 1:
					b.Y = l
				}
				continue
			}
			n := m[a]
			if n != "" {
				l := &ast.Ident{Name: n}
				switch i {
				case 0:
					b.X = l
				case 1:
					b.Y = l
				}
				continue
			}
			if i != 0 {
				log.Logger().Error().
					Str("op", op).
					Msgf("not found IdentExpr %#v in %#v", a, x)
				continue
			}
			b.X = pb
		}

	}

	return b, nil
}
