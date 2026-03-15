package archlint

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// GinLeakAnalyzer ensures *gin.Context is not used in application layer UseCases.
var GinLeakAnalyzer = &analysis.Analyzer{
	Name:     "ginleak",
	Doc:      "ensures *gin.Context is not used in application layer UseCases",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runGinLeakCheck,
}

func runGinLeakCheck(pass *analysis.Pass) (any, error) {
	if !strings.Contains(pass.Pkg.Path(), "/application") {
		return nil, nil
	}

	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		fn, ok := node.(*ast.FuncDecl)
		if !ok || fn.Type == nil || fn.Type.Params == nil {
			return
		}

		for _, field := range fn.Type.Params.List {
			for _, name := range field.Names {
				obj := pass.TypesInfo.Defs[name]
				if obj == nil {
					continue
				}

				if isGinContext(obj.Type()) {
					pass.Reportf(field.Pos(), "Arch violation: Application layer function %s uses *gin.Context. Use context.Context instead.", fn.Name.Name)
				}
			}
			// anonymous parameters
			if len(field.Names) == 0 {
				if isGinContext(pass.TypesInfo.TypeOf(field.Type)) {
					pass.Reportf(field.Pos(), "Arch violation: Application layer function %s uses *gin.Context. Use context.Context instead.", fn.Name.Name)
				}
			}
		}
	})

	return nil, nil
}

func isGinContext(t types.Type) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return false
	}

	return named.Obj().Pkg().Path() == "github.com/gin-gonic/gin" && named.Obj().Name() == "Context"
}
