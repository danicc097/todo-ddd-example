package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/danicc097/todo-ddd-example/tools/archlint"
)

func main() {
	multichecker.Main(
		archlint.HandlerAnalyzer,
		archlint.GinLeakAnalyzer,
		archlint.CQRSPurityAnalyzer,
		archlint.DTOBleedAnalyzer,
		archlint.DomainImmutabilityAnalyzer,
		archlint.EventRegistrationAnalyzer,
		archlint.InfrastructureLeakAnalyzer,
		archlint.EventMapperAnalyzer,
	)
}
