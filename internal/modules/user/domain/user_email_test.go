package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"invalid format", "testexample.com", true},
		{"empty", "", true},
		{"whitespace", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewUserEmail(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, strings.ToLower(strings.TrimSpace(tt.input)), email.String())
			}
		})
	}
}
