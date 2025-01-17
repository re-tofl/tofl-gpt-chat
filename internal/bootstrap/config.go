package bootstrap

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	ServerPort              string `mapstructure:"SERVER_PORT"`
	TGBotToken              string `mapstructure:"TG_BOT_TOKEN"`
	LLMURL                  string `mapstructure:"LLM_URL"`
	ParserURL               string `mapstructure:"PARSER_URL"`
	FormalURL               string `mapstructure:"FORMAL_URL"`
	YandexKey               string `mapstructure:"YANDEX_KEY"`
	YandexTranslateUrl      string `mapstructure:"YANDEX_TRANSLATE_URL"`
	YandexTranslateFolderId string `mapstructure:"YANDEX_TRANSLATE_FOLDER_ID"`
}

func Setup(cfgPath string) (*Config, error) {
	viper.SetConfigFile(cfgPath)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if os.IsNotExist(err) {
		fmt.Print("config file not found, skipping")
	} else if err != nil {
		return nil, err
	}

	return UnmarshalConfig(), nil
}

func UnmarshalConfig() *Config {
	return &Config{
		ServerPort:              GetStringOr("SERVER_PORT", "8080"),
		TGBotToken:              viper.GetString("TG_BOT_TOKEN"),
		LLMURL:                  viper.GetString("LLM_URL"),
		ParserURL:               viper.GetString("PARSER_URL"),
		FormalURL:               viper.GetString("FORMAL_URL"),
		YandexKey:               viper.GetString("YANDEX_KEY"),
		YandexTranslateUrl:      viper.GetString("YANDEX_TRANSLATE_URL"),
		YandexTranslateFolderId: viper.GetString("YANDEX_TRANSLATE_FOLDER_ID"),
	}
}

func GetStringOr(key string, defaultValue string) string {
	if viper.IsSet(key) {
		return viper.GetString(key)
	}

	return defaultValue
}

func GetIntOr(key string, defaultValue int) int {
	if viper.IsSet(key) {
		return viper.GetInt(key)
	}

	return defaultValue
}
