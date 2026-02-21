package internal

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	configLock = &sync.Mutex{}
	Config     *AppConfig
)

type AppEnv string

const (
	AppEnvDev  AppEnv = "development"
	AppEnvProd AppEnv = "production"
	AppEnvCI   AppEnv = "ci"
)

func (e *AppEnv) UnmarshalText(text []byte) error {
	value := string(text)
	switch value {
	case string(AppEnvDev), string(AppEnvProd), string(AppEnvCI):
		*e = AppEnv(value)
	default:
		return fmt.Errorf("invalid value for AppEnv: %v", value)
	}

	return nil
}

type PostgresConfig struct {
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASS"`
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	DBName   string `mapstructure:"DB_NAME"`
}

type RedisConfig struct {
	Addr string `mapstructure:"REDIS_ADDR"`
}

type RabbitMQConfig struct {
	URL string `mapstructure:"RABBITMQ_URL"`
}

type OTELConfig struct {
	Endpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}

type AppConfig struct {
	Postgres     PostgresConfig `mapstructure:",squash"`
	Redis        RedisConfig    `mapstructure:",squash"`
	RabbitMQ     RabbitMQConfig `mapstructure:",squash"`
	OTEL         OTELConfig     `mapstructure:",squash"`
	LogLevel     string         `mapstructure:"LOG_LEVEL"`
	Env          AppEnv         `mapstructure:"ENV"`
	Port         string         `mapstructure:"PORT"`
	MFAMasterKey string         `mapstructure:"MFA_MASTER_KEY"`
}

// NewAppConfig initializes the global Config variable.
func NewAppConfig() error {
	configLock.Lock()
	defer configLock.Unlock()

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	Config = cfg

	return nil
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*AppConfig, error) {
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	bindEnvs(v, AppConfig{})

	v.SetDefault("LOG_LEVEL", "INFO")
	v.SetDefault("ENV", "development")

	cfg := &AppConfig{}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("viper unmarshal: %w", err)
	}

	return cfg, nil
}

func bindEnvs(v *viper.Viper, iface any, parts ...string) {
	ifv := reflect.ValueOf(iface)

	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		vfield := ifv.Field(i)
		tfield := ift.Field(i)

		tv, ok := tfield.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}

		if tv == ",squash" {
			bindEnvs(v, vfield.Interface(), parts...)
			continue
		}

		switch vfield.Kind() {
		case reflect.Struct:
			bindEnvs(v, vfield.Interface(), append(parts, tv)...)
		default:
			v.BindEnv(strings.Join(append(parts, tv), "."))
		}
	}
}
