package archlint

import (
	"strings"

	"golang.org/x/tools/go/analysis"
)

var InfrastructureLeakAnalyzer = &analysis.Analyzer{
	Name: "infraleak",
	Doc:  "ensures application layer does not import infrastructure packages directly",
	Run:  runInfrastructureLeakCheck,
}

func runInfrastructureLeakCheck(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	if !strings.Contains(pkgPath, "/modules/") || !strings.Contains(pkgPath, "/application") {
		return nil, nil
	}

	if isTestOnly(pass) {
		return nil, nil
	}

	for _, imp := range pass.Pkg.Imports() {
		impPath := imp.Path()
		if strings.Contains(impPath, "/infrastructure") {
			pass.Reportf(pass.Files[0].Package, "Arch violation: Application package %s imports infrastructure package %s. Depend on domain ports instead.", pkgPath, impPath)
		}
	}

	return nil, nil
}

func isTestOnly(pass *analysis.Pass) bool {
	for _, f := range pass.Files {
		name := pass.Fset.File(f.Pos()).Name()
		if !strings.HasSuffix(name, "_test.go") {
			return false
		}
	}

	return true
}
