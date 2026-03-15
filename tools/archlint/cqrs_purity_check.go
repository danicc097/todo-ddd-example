package archlint

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// CQRSPurityAnalyzer ensures QueryServices are read-only and don't use UnitOfWork.
var CQRSPurityAnalyzer = &analysis.Analyzer{
	Name:     "cqrsarch",
	Doc:      "ensures QueryServices are read-only and don't use UnitOfWork",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runCQRSPurityCheck,
}
var allowedVerbs = []string{"Get", "List", "Find", "Search", "Count", "GetAll"}

func runCQRSPurityCheck(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{(*ast.TypeSpec)(nil)}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		n := node.(*ast.TypeSpec)
		if iface, ok := n.Type.(*ast.InterfaceType); ok {
			if strings.HasSuffix(n.Name.Name, "QueryService") {
				checkInterface(pass, iface, n.Name.Name)
			}
		} else if st, ok := n.Type.(*ast.StructType); ok {
			if strings.HasSuffix(n.Name.Name, "QueryService") || strings.Contains(strings.ToLower(n.Name.Name), "query") {
				checkStructFields(pass, st, n.Name.Name)
			}
		}
	})

	return nil, nil
}

func checkInterface(pass *analysis.Pass, iface *ast.InterfaceType, ifaceName string) {
	if iface.Methods == nil {
		return
	}

	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			continue
		}

		methodName := method.Names[0].Name
		allowed := false

		for _, verb := range allowedVerbs {
			if strings.HasPrefix(methodName, verb) {
				allowed = true
				break
			}
		}

		if !allowed {
			pass.Reportf(method.Pos(), "Arch violation: QueryService interface %s has non-query method %s. Queries must only use allowed verbs: %v.", ifaceName, methodName, allowedVerbs)
		}
	}
}

func checkStructFields(pass *analysis.Pass, st *ast.StructType, structName string) {
	if st.Fields == nil {
		return
	}

	for _, field := range st.Fields.List {
		t := pass.TypesInfo.TypeOf(field.Type)
		if t == nil {
			continue
		}

		if isUnitOfWork(t) {
			pass.Reportf(field.Pos(), "Arch violation: %s cannot have UnitOfWork as a field. Queries must be read-only.", structName)
		}
	}
}

func isUnitOfWork(t types.Type) bool {
	s := t.String()
	return strings.Contains(s, "shared/application.UnitOfWork")
}
