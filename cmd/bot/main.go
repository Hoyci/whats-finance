package main

import (
	"fmt"
	"os"

	"github.com/hoyci/whats-finance/internal/config"
	"github.com/hoyci/whats-finance/internal/handler"
	"github.com/hoyci/whats-finance/internal/processor"
	googlesheets "github.com/hoyci/whats-finance/pkg/google-sheets"
	"github.com/hoyci/whats-finance/pkg/whatsapp"
)

func main() {
	cfg := config.GetConfig()
	if cfg == nil {
		fmt.Println("Falha ao carregar as configurações. Verifique o arquivo .env.")
		os.Exit(1)
	}

	whatsappClient := whatsapp.GetInstance(cfg.DBPath)

	sheetsService, err := googlesheets.NewSheetsService(cfg.CredentialsJSON, cfg.GoogleSheetID)
	if err != nil {
		fmt.Printf("Falha ao inicializar o cliente do Google Sheets: %v\n", err)
		os.Exit(1)
	}

	chatGPTProcessor := processor.NewChatGPTProcessor(cfg.ChatGPTApiKey)

	botHandler := handler.NewBotHandler(
		whatsappClient,
		chatGPTProcessor,
		sheetsService,
		cfg.SheetName,
		cfg.PhoneNumber,
	)

	botHandler.InitializeBot()
	fmt.Println("Bot iniciado e escutando mensagens...")
}
