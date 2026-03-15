package archlint

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

//nolint:gochecknoglobals // required by go/analysis
var HandlerAnalyzer = &analysis.Analyzer{
	Name:     "handlerarch",
	Doc:      "ensures HTTP handlers do not call Repositories directly",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runHandlerArch,
}

//nolint:nilnil // no meaningful result to return
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
		if !ok || funcDecl.Body == nil {
			return
		}

		if isHTTPHandler(pass, funcDecl) {
			inspectBody(pass, funcDecl.Body, funcDecl.Name.Name)
		}
	})

	return nil, nil
}

func isHTTPHandler(pass *analysis.Pass, funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil {
		return false
	}

	for _, field := range funcDecl.Type.Params.List {
		t := pass.TypesInfo.TypeOf(field.Type)
		if isGinContext(t) {
			return true
		}
	}

	return false
}

func inspectBody(pass *analysis.Pass, body *ast.BlockStmt, funcName string) {
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

		checkViolation(pass, obj, call, funcName, sel.Sel.Name)

		return true
	})
}

func checkViolation(pass *analysis.Pass, obj types.Object, call *ast.CallExpr, funcName, methodName string) {
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
		pass.Reportf(call.Pos(), "Arch violation: HTTP handler %s calls %s.%s directly. Handlers must route through Application UseCases.", funcName, iface, methodName)
	}
}
