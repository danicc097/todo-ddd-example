package domain

import (
	"fmt"
	"strings"
)

type AuditOperation string

const (
	OpCreate AuditOperation = "CREATE"
	OpUpdate AuditOperation = "UPDATE"
	OpUpsert AuditOperation = "UPSERT"
	OpDelete AuditOperation = "DELETE"
	OpRead   AuditOperation = "READ"
)

func (o AuditOperation) IsValid() error {
	switch o {
	case OpCreate, OpUpdate, OpDelete, OpRead, OpUpsert:
		return nil
	default:
		return fmt.Errorf("invalid audit operation: %s", o)
	}
}

func (o AuditOperation) String() string {
	return string(o)
}

func ParseAuditOperation(s string) (AuditOperation, error) {
	op := AuditOperation(strings.ToUpper(strings.TrimSpace(s)))
	if err := op.IsValid(); err != nil {
		return "", err
	}

	return op, nil
}
