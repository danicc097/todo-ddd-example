package domain

// Validatable defines objects that can be validated.
type Validatable interface {
	Validate() error
}
