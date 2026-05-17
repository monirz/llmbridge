package config

import (
	"log"

	"github.com/spf13/viper"
)

type Provider struct {
	ID    string `mapstructure:"id"    json:"id"`
	Name  string `mapstructure:"name"  json:"name"`
	Model string `mapstructure:"model" json:"model"`
	Badge string `mapstructure:"badge" json:"badge"`
	Icon  string `mapstructure:"icon"  json:"icon"`
	Color string `mapstructure:"color" json:"color"`
}

type Config struct {
	BifrostURL   string     `mapstructure:"bifrost_url"`
	DefaultModel string     `mapstructure:"default_model"`
	Providers    []Provider `mapstructure:"providers"`
}

func Load() *Config {
	viper.SetConfigName("provider_config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.BindEnv("bifrost_url", "BIFROST_URL")
	viper.BindEnv("default_model", "MODEL")

	viper.SetDefault("bifrost_url", "http://localhost:8080/v1/chat/completions")
	viper.SetDefault("default_model", "ollama/qwen2.5:3b")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("no provider_config.yaml found, using defaults and env vars")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	return &cfg
}
