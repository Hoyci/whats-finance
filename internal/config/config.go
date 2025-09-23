package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var (
	config *Config
	once   sync.Once
)

type Config struct {
	ChatGPTApiKey   string
	GoogleSheetID   string
	PhoneNumber     string
	SheetName       string
	CredentialsJSON string
	DBPath          string
}

func GetConfig() *Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			fmt.Println("Nenhum arquivo .env encontrado ou erro ao carregar.")
		}

		config = &Config{
			ChatGPTApiKey:   os.Getenv("CHATGPT_API_KEY"),
			GoogleSheetID:   os.Getenv("GOOGLE_SHEET_ID"),
			PhoneNumber:     os.Getenv("PHONE_NUMBER"),
			SheetName:       os.Getenv("SHEET_NAME"),
			CredentialsJSON: os.Getenv("CREDENTIALS_JSON"),
			DBPath:          os.Getenv("DB_PATH"),
		}

		if config.ChatGPTApiKey == "" ||
			config.GoogleSheetID == "" ||
			config.PhoneNumber == "" ||
			config.SheetName == "" ||
			config.CredentialsJSON == "" {
			log.Fatal("Variável de ambiente não foi definida.")
		}
	})

	return config
}
