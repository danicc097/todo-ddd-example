package domain

// AppConfig provides application-level configuration to the domain.
type AppConfig interface {
	DisplayName() string
}
