package archlint

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// EventRegistrationAnalyzer ensures Domain Events are passed to RecordEvent().
var EventRegistrationAnalyzer = &analysis.Analyzer{
	Name:     "eventregistration",
	Doc:      "ensures Domain Events are passed to RecordEvent() to ensure they are persisted",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runEventRegistrationCheck,
}

func runEventRegistrationCheck(pass *analysis.Pass) (any, error) {
	if !strings.Contains(pass.Pkg.Path(), "/domain") {
		return nil, nil
	}

	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Body == nil {
			return
		}

		if strings.HasSuffix(pass.Fset.File(fn.Pos()).Name(), "_test.go") {
			return
		}

		eventVars := make(map[types.Object]ast.Node)
		recordedVars := make(map[types.Object]bool)

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.AssignStmt:
				for _, rhs := range x.Rhs {
					if isDomainEvent(pass, rhs) {
						for _, lhs := range x.Lhs {
							if ident, ok := lhs.(*ast.Ident); ok {
								if obj := pass.TypesInfo.ObjectOf(ident); obj != nil {
									eventVars[obj] = n
								}
							}
						}
					}
				}
			case *ast.CallExpr:
				if isRecordEventCall(pass, x) {
					for _, arg := range x.Args {
						if ident, ok := arg.(*ast.Ident); ok {
							if obj := pass.TypesInfo.ObjectOf(ident); obj != nil {
								recordedVars[obj] = true
							}
						}
					}
				}
			}

			return true
		})

		for obj, node := range eventVars {
			if !recordedVars[obj] {
				pass.Reportf(node.Pos(), "Arch violation: Domain event variable %s is created but not passed to RecordEvent().", obj.Name())
			}
		}
	})

	return nil, nil
}

func isDomainEvent(pass *analysis.Pass, expr ast.Expr) bool {
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return false
	}

	s := t.String()

	return strings.Contains(s, "domain.") && strings.HasSuffix(s, "Event")
}

func isRecordEventCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		if ident, ok := call.Fun.(*ast.Ident); ok {
			return ident.Name == "RecordEvent"
		}

		return false
	}

	return sel.Sel.Name == "RecordEvent"
}
