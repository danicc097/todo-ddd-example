package archlint

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// DTOBleedAnalyzer ensures internal/generated/db types are not used in application or http layers.
var DTOBleedAnalyzer = &analysis.Analyzer{
	Name:     "dtobleed",
	Doc:      "ensures internal/generated/db types are not used in application or http layers",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runDTOBleedCheck,
}

func runDTOBleedCheck(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	if !strings.Contains(pkgPath, "/application") && !strings.Contains(pkgPath, "/infrastructure/http") {
		return nil, nil
	}

	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil), (*ast.TypeSpec)(nil)}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		switch n := node.(type) {
		case *ast.FuncDecl:
			checkFuncSignature(pass, n)
		case *ast.TypeSpec:
			if st, ok := n.Type.(*ast.StructType); ok {
				checkStructDTO(pass, st, n.Name.Name)
			}
		}
	})

	return nil, nil
}

func checkFuncSignature(pass *analysis.Pass, fn *ast.FuncDecl) {
	if fn.Type == nil {
		return
	}

	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			if isDBType(pass.TypesInfo.TypeOf(field.Type)) {
				pass.Reportf(field.Pos(), "Arch violation: Function %s uses type from internal/generated/db in parameters. Use a Domain entity or API DTO instead.", fn.Name.Name)
			}
		}
	}

	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			if isDBType(pass.TypesInfo.TypeOf(field.Type)) {
				pass.Reportf(field.Pos(), "Arch violation: Function %s returns type from internal/generated/db. Use a Domain entity or API DTO instead.", fn.Name.Name)
			}
		}
	}
}

func checkStructDTO(pass *analysis.Pass, st *ast.StructType, structName string) {
	if st.Fields == nil {
		return
	}

	for _, field := range st.Fields.List {
		if isDBType(pass.TypesInfo.TypeOf(field.Type)) {
			pass.Reportf(field.Pos(), "Arch violation: Struct %s has field using type from internal/generated/db. Use a Domain entity or API DTO instead.", structName)
		}
	}
}

func isDBType(t types.Type) bool {
	if t == nil {
		return false
	}

	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	if slice, ok := t.(*types.Slice); ok {
		t = slice.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return false
	}

	return strings.HasSuffix(named.Obj().Pkg().Path(), "/internal/generated/db")
}
