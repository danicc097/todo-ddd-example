package domain

import (
	"fmt"
	"strings"
)

type AuditAggregateType string

const (
	AggWorkspace AuditAggregateType = "WORKSPACE"
	AggTodo      AuditAggregateType = "TODO"
	AggUser      AuditAggregateType = "USER"
)

func (a AuditAggregateType) IsValid() error {
	switch a {
	case AggWorkspace, AggTodo, AggUser:
		return nil
	default:
		return fmt.Errorf("invalid aggregate type: %s", a)
	}
}

func (a AuditAggregateType) String() string {
	return string(a)
}

func ParseAuditAggregateType(s string) (AuditAggregateType, error) {
	agg := AuditAggregateType(strings.ToUpper(strings.TrimSpace(s)))
	if err := agg.IsValid(); err != nil {
		return "", err
	}

	return agg, nil
}
