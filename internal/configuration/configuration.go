package configuration

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type MissingEnvConfigError struct {
	env string
	err error
}

func (mece MissingEnvConfigError) Error() string {
	return fmt.Sprintf("missing config %s: %v", mece.env, mece.err)
}

type MissingBaseConfigError struct {
	err error
}

func (mbce MissingBaseConfigError) Error() string {
	return fmt.Sprintf("missing base config: %v", mbce.err)
}

func GetConfig(currEnv string) (*Config, error) {
	// config
	var cfg Config

	if err := cleanenv.ReadConfig("./config/base.yaml", &cfg); err != nil {
		return nil, MissingBaseConfigError{err: err}
	}

	if err := cleanenv.ReadConfig(fmt.Sprintf("./config/%s.yaml", currEnv), &cfg); err != nil {
		return &cfg, MissingEnvConfigError{env: currEnv, err: err}
	}

	return &cfg, nil
}
