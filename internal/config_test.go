package internal

import (
	"fmt"
	"testing"

	"github.com/danicc097/todo-ddd-example/internal/utils/pointers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppConfig(t *testing.T) {
	type nestedCfg struct {
		Name string `env:"TEST_CFG_NAME"`
	}
	type cfg struct {
		NestedCfg       nestedCfg
		Length          int     `env:"TEST_CFG_LEN"`
		OptionalLength  *int    `env:"TEST_CFG_OPT_LEN"`
		OptionalString  *string `env:"TEST_CFG_STRING_PTR"`
		BoolWithDefault bool    `env:"TEST_CFG_BOOL_DEFAULT,false"`
	}

	type params struct {
		name        string
		want        *cfg
		errContains string
		environ     map[string]string
	}

	tests := []params{
		{
			name:    "correct env",
			want:    &cfg{NestedCfg: nestedCfg{Name: "name"}, Length: 10, OptionalLength: pointers.New(40), BoolWithDefault: true},
			environ: map[string]string{"TEST_CFG_NAME": "name", "TEST_CFG_LEN": "10", "TEST_CFG_OPT_LEN": "40", "TEST_CFG_BOOL_DEFAULT": "true"},
		},
		{
			name:    "missing pointer fields is ok",
			want:    &cfg{NestedCfg: nestedCfg{Name: "name"}, Length: 10, BoolWithDefault: false},
			environ: map[string]string{"TEST_CFG_NAME": "name", "TEST_CFG_LEN": "10"},
		},
		{
			name:        "bad int conversion",
			environ:     map[string]string{"TEST_CFG_NAME": "name", "TEST_CFG_LEN": "aaa"},
			errContains: `could not convert TEST_CFG_LEN to int`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.environ {
				t.Setenv(k, v)
			}

			c := &cfg{}
			err := loadEnvToConfig(c)
			if tc.errContains != "" {
				require.ErrorContains(t, err, tc.errContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, c)
		})
	}
}

type MyEnum string

const (
	Value1 MyEnum = "Value1"
	Value2 MyEnum = "Value2"
)

func (e *MyEnum) Decode(value string) error {
	switch value {
	case "1":
		*e = Value1
	case "2":
		*e = Value2
	default:
		return fmt.Errorf("invalid value for MyEnum: %v", value)
	}
	return nil
}

func TestEnumDecoderConfig(t *testing.T) {
	type cfg struct {
		Enum         MyEnum  `env:"TEST_CFG_ENUM"`
		OptionalEnum *MyEnum `env:"TEST_CFG_ENUM_OPT"`
	}

	t.Run("correct decoding", func(t *testing.T) {
		t.Setenv("TEST_CFG_ENUM", "1")
		t.Setenv("TEST_CFG_ENUM_OPT", "2")
		c := &cfg{}
		err := loadEnvToConfig(c)
		require.NoError(t, err)
		assert.Equal(t, Value1, c.Enum)
		assert.Equal(t, Value2, *c.OptionalEnum)
	})
}
