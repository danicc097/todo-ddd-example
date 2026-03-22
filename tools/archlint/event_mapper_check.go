package archlint

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var EventMapperAnalyzer = &analysis.Analyzer{
	Name:     "eventmapper",
	Doc:      "ensures uow.Collect() is called with a non-nil EventMapper to prevent silent event loss",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runEventMapperCheck,
}

func runEventMapperCheck(pass *analysis.Pass) (any, error) {
	if !strings.Contains(pass.Pkg.Path(), "/infrastructure/postgres") {
		return nil, nil
	}

	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		call := node.(*ast.CallExpr)

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Collect" {
			return
		}

		obj := pass.TypesInfo.Uses[sel.Sel]
		if obj == nil {
			return
		}

		if !isUoWCollect(obj) {
			return
		}

		if len(call.Args) < 2 {
			return
		}

		// Collect(ctx, mapper, agg)
		mapperArg := call.Args[1]
		if ident, ok := mapperArg.(*ast.Ident); ok && ident.Name == "nil" {
			pass.Reportf(call.Pos(), "Arch violation: uow.Collect() called with nil EventMapper. Domain events will be silently lost. Provide a concrete EventMapper implementation.")
		}
	})

	return nil, nil
}

func isUoWCollect(obj types.Object) bool {
	sig, ok := obj.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return false
	}

	recvType := sig.Recv().Type().String()

	return strings.Contains(recvType, "UnitOfWork")
}
