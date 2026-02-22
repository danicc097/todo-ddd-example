package http

var SensitiveFields = map[string]struct{}{
	"password": {},
	"code":     {},
}
