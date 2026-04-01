package cmd

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

func LoadConfig[T any]() (*T, error) {
	var cfg T
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	validate := validator.New()
	err := validate.Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)

	}

	return &cfg, nil
}
