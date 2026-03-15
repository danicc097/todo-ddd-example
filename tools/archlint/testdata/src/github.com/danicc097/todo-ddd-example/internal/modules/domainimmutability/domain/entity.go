package domain

// BadEntity has exported fields -- should be flagged.
type BadEntity struct {
	ID     string // want "Arch violation: Domain struct BadEntity has exported field ID. Use methods to enforce invariants."
	Status string // want "Arch violation: Domain struct BadEntity has exported field Status. Use methods to enforce invariants."
}

// GoodEntity has only unexported fields.
type GoodEntity struct {
	id     string
	status string
}

func (g *GoodEntity) ID() string     { return g.id }
func (g *GoodEntity) Status() string { return g.status }

// ReconstituteBadEntityArgs ends with "Args" -- exported fields are permitted.
type ReconstituteBadEntityArgs struct {
	ID     string
	Status string
}

// BadEntityCreatedEvent ends with "Event" -- exported fields are permitted.
type BadEntityCreatedEvent struct {
	ID         string
	OccurredAt string
}
