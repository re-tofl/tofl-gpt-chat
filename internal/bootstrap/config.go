package bootstrap

import (
	"github.com/spf13/viper"
)

type Config struct {
	ServerPort  string `mapstructure:"SERVER_PORT"`
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	TGKey       string `mapstructure:"TG_KEY"`
}

func Setup(cfgPath string) (*Config, error) {
	viper.SetConfigFile(cfgPath)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var cfg Config

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
