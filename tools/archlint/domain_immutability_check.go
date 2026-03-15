package archlint

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// DomainImmutabilityAnalyzer ensures domain entities and VOs have no exported fields.
var DomainImmutabilityAnalyzer = &analysis.Analyzer{
	Name:     "domainimmutability",
	Doc:      "ensures domain entities and VOs have no exported fields to enforce method-based mutations",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runDomainImmutabilityCheck,
}

func runDomainImmutabilityCheck(pass *analysis.Pass) (any, error) {
	if !strings.Contains(pass.Pkg.Path(), "/domain") {
		return nil, nil
	}

	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{(*ast.TypeSpec)(nil)}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		ts := node.(*ast.TypeSpec)
		if strings.HasSuffix(ts.Name.Name, "Args") || strings.HasSuffix(ts.Name.Name, "Event") {
			return
		}

		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return
		}

		if st.Fields == nil {
			return
		}

		for _, field := range st.Fields.List {
			for _, name := range field.Names {
				if ast.IsExported(name.Name) {
					pass.Reportf(name.Pos(), "Arch violation: Domain struct %s has exported field %s. Use methods to enforce invariants.", ts.Name.Name, name.Name)
				}
			}
		}
	})

	return nil, nil
}
