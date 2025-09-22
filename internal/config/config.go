package config

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

var (
	config *Config
	once   sync.Once
)

type Config struct {
	ChatGPTApiKey string `mapstructure:"CHATGPT_API_KEY"`
	GoogleSheetID string `mapstructure:"GOOGLE_SHEET_ID"`
	PhoneNumber   string `mapstructure:"PHONE_NUMBER"`
	SheetName     string `mapstructure:"SHEET_NAME"`
}

func GetConfig() *Config {
	once.Do(func() {
		viper.SetConfigName(".env")
		viper.SetConfigType("env")
		viper.AddConfigPath(".")
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("error reading config file, %s", err)
		}

		if err := viper.Unmarshal(&config); err != nil {
			log.Fatalf("error unmarshalling config, %s", err)
		}
	})

	return config
}
