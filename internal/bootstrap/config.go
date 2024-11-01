package bootstrap

import (
	"github.com/spf13/viper"
)

type Config struct {
	ServerPort              string `mapstructure:"SERVER_PORT"`
	DatabaseURL             string `mapstructure:"DATABASE_URL"`
	TGBotToken              string `mapstructure:"TG_BOT_TOKEN"`
	YandexKey               string `mapstructure:"YANDEX_KEY"`
	YandexTranslateUrl      string `mapstructure:"YANDEX_TRANSLATE_URL"`
	YandexTranslateFolderId string `mapstructure:"YANDEX_TRANSLATE_FOLDER_ID"`
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
