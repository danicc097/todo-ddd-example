package domain

type AggregateType string

const (
	AggWorkspace AggregateType = "WORKSPACE"
	AggTodo      AggregateType = "TODO"
	AggUser      AggregateType = "USER"
	AggTag       AggregateType = "TAG"
)

func (a AggregateType) String() string {
	return string(a)
}
