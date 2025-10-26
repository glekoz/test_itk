package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	Port        string `mapstructure:"ITKAPP_PORT"`
	Host        string `mapstructure:"HOST"`
	CacheTTL    int    `mapstructure:"CACHE_TTL"`
}

func MustLoad() *Config {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/itkapp/")

	err := viper.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var config Config

	// Преобразование в структуру
	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %w", err))
	}

	if config.CacheTTL < 1 {
		panic(fmt.Errorf("cache ttl must be greater than 0"))
	}

	return &config
}
