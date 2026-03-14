package archlint

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

//nolint:gochecknoglobals
var HandlerAnalyzer = &analysis.Analyzer{
	Name:     "handlerarch",
	Doc:      "ensures HTTP handlers do not call Repositories directly",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runHandlerArch,
}

//nolint:nilnil
func runHandlerArch(pass *analysis.Pass) (any, error) {
	if !strings.Contains(pass.Pkg.Path(), "/infrastructure/http") {
		return nil, nil
	}

	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		funcDecl, ok := node.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil || funcDecl.Body == nil {
			return
		}

		recvName := getReceiverName(funcDecl)
		if !strings.HasSuffix(recvName, "Handler") {
			return
		}

		inspectBody(pass, funcDecl.Body, recvName)
	})

	return nil, nil
}

func getReceiverName(funcDecl *ast.FuncDecl) string {
	if len(funcDecl.Recv.List) == 0 {
		return ""
	}

	switch exp := funcDecl.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if ident, ok := exp.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return exp.Name
	}

	return ""
}

func inspectBody(pass *analysis.Pass, body *ast.BlockStmt, recvName string) {
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		obj := pass.TypesInfo.Uses[sel.Sel]
		if obj == nil {
			return true
		}

		checkViolation(pass, obj, call, recvName, sel.Sel.Name)

		return true
	})
}

func checkViolation(pass *analysis.Pass, obj types.Object, call *ast.CallExpr, recvName, methodName string) {
	sig, ok := obj.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return
	}

	recvType := sig.Recv().Type()

	if ptr, isPtr := recvType.(*types.Pointer); isPtr {
		recvType = ptr.Elem()
	}

	named, ok := recvType.(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return
	}

	iface := named.Obj().Name()
	pkgPath := named.Obj().Pkg().Path()

	if strings.Contains(pkgPath, "/domain") && strings.HasSuffix(iface, "Repository") {
		pass.Reportf(call.Pos(), "Arch violation: %s calls %s.%s directly. Handlers must route through Application UseCases.", recvName, iface, methodName)
	}
}
