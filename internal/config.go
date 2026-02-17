package internal

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
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

func (e *AppEnv) Decode(value string) error {
	switch value {
	case string(AppEnvDev), string(AppEnvProd), string(AppEnvCI):
		*e = AppEnv(value)
	default:
		return fmt.Errorf("invalid value for AppEnv: %v", value)
	}

	return nil
}

type PostgresConfig struct {
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASS"`
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	DBName   string `env:"DB_NAME"`
}

type RedisConfig struct {
	Addr string `env:"REDIS_ADDR"`
}

type RabbitMQConfig struct {
	URL string `env:"RABBITMQ_URL"`
}

type OTELConfig struct {
	Endpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}

type AppConfig struct {
	Postgres     PostgresConfig
	Redis        RedisConfig
	RabbitMQ     RabbitMQConfig
	OTEL         OTELConfig
	LogLevel     string `env:"LOG_LEVEL,INFO"`
	Env          AppEnv `env:"ENV,development"`
	Port         string `env:"PORT"`
	MFAMasterKey string `env:"MFA_MASTER_KEY"`
}

func NewAppConfig() error {
	configLock.Lock()
	defer configLock.Unlock()

	cfg := &AppConfig{}

	if err := loadEnvToConfig(cfg); err != nil {
		return fmt.Errorf("loadEnvToConfig: %w", err)
	}

	Config = cfg

	return nil
}

var decoderType = reflect.TypeFor[Decoder]()

type Decoder interface {
	Decode(value string) error
}

func loadEnvToConfig(config any) error {
	cfg := reflect.ValueOf(config)

	if cfg.Kind() == reflect.Pointer {
		cfg = cfg.Elem()
	}

	for idx := 0; idx < cfg.NumField(); idx++ {
		fType := cfg.Type().Field(idx)
		fld := cfg.Field(idx)

		if fld.Kind() == reflect.Struct {
			if !fld.CanInterface() {
				continue
			}

			if err := loadEnvToConfig(fld.Addr().Interface()); err != nil {
				return fmt.Errorf("nested struct %q env loading: %w", fType.Name, err)
			}
		}

		if !fld.CanSet() {
			continue
		}

		if envtag, ok := fType.Tag.Lookup("env"); ok && len(envtag) > 0 {
			isPtr := fld.Kind() == reflect.Pointer

			var ptr reflect.Type
			if isPtr {
				ptr = fld.Type()
			} else {
				ptr = reflect.PtrTo(fType.Type)
			}

			if ptr.Implements(decoderType) {
				envvar, _ := splitEnvTag(envtag)

				val, _ := os.LookupEnv(envvar)
				if val == "" && isPtr {
					continue
				}

				var (
					decoder Decoder
					ok      bool
				)

				if isPtr {
					decoder, ok = reflect.New(ptr.Elem()).Interface().(Decoder)
				} else {
					decoder, ok = fld.Addr().Interface().(Decoder)
				}

				if !ok {
					return fmt.Errorf("%q: could not find Decoder method", ptr.Elem())
				}

				if err := setDecoderValue(decoder, fType.Tag.Get("env"), fld); err != nil {
					return fmt.Errorf("could not decode %q: %w", fType.Name, err)
				}

				if isPtr {
					fld.Set(reflect.ValueOf(decoder))
				} else {
					fld.Set(reflect.ValueOf(decoder).Elem())
				}

				continue
			}

			if err := setEnvToField(envtag, fld); err != nil {
				return fmt.Errorf("could not set %q to %q: %w", envtag, fType.Name, err)
			}
		}
	}

	return nil
}

func setDecoderValue(decoder Decoder, envTag string, field reflect.Value) error {
	envvar, defaultVal := splitEnvTag(envTag)
	val, present := os.LookupEnv(envvar)

	if !present && field.Kind() != reflect.Pointer {
		if defaultVal == "" {
			return fmt.Errorf("%s is not set but required", envvar)
		}

		val = defaultVal
	}

	var isPtr bool

	kind := field.Kind()

	if kind == reflect.Pointer {
		kind = field.Type().Elem().Kind()
		isPtr = true
	}

	if val == "" && isPtr && kind != reflect.String {
		return nil
	}

	return decoder.Decode(val)
}

func splitEnvTag(s string) (string, string) {
	x := strings.Split(s, ",")
	if len(x) == 1 {
		return x[0], ""
	}

	return x[0], x[1]
}

func setEnvToField(envTag string, field reflect.Value) error {
	envvar, defaultVal := splitEnvTag(envTag)
	val, present := os.LookupEnv(envvar)

	if !present && field.Kind() != reflect.Pointer {
		if defaultVal == "" {
			return fmt.Errorf("%s is not set but required", envvar)
		}

		val = defaultVal
	}

	var isPtr bool

	kind := field.Kind()
	if kind == reflect.Pointer {
		kind = field.Type().Elem().Kind()
		isPtr = true
	}

	if val == "" && isPtr && kind != reflect.String {
		return nil
	}

	switch kind {
	case reflect.String:
		if !present && isPtr {
			setVal[*string](false, field, nil)
			return nil
		}

		setVal(isPtr, field, val)
	case reflect.Int:
		v, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("could not convert %s to int: %w", envvar, err)
		}

		setVal(isPtr, field, v)
	case reflect.Bool:
		v, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("could not convert %s to bool: %w", envvar, err)
		}

		setVal(isPtr, field, v)
	default:
		return fmt.Errorf("unsupported type for env tag %q: %T", envvar, field.Interface())
	}

	return nil
}

func setVal[T any](isPtr bool, field reflect.Value, v T) {
	if isPtr {
		field.Set(reflect.ValueOf(&v))
	} else {
		field.Set(reflect.ValueOf(v))
	}
}
