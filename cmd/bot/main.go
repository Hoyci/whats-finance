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

	waClient, err := whatsapp.InitializeClient(cfg.DBPath)
	if err != nil {
		fmt.Printf("Falha ao inicializar o cliente do WhatsApp: %v\n", err)
		os.Exit(1)
	}

	sheetsService, err := googlesheets.NewSheetsService(cfg.CredentialsJSON, cfg.GoogleSheetID)
	if err != nil {
		fmt.Printf("Falha ao inicializar o cliente do Google Sheets: %v\n", err)
		os.Exit(1)
	}

	chatGPTProcessor := processor.NewChatGPTProcessor(cfg.ChatGPTApiKey)

	msgHandler, err := handler.NewMessageHandler(
		waClient,
		cfg.PhoneNumber,
		chatGPTProcessor,
		sheetsService,
		cfg.SheetName,
	)
	if err != nil {
		fmt.Printf("Falha ao criar o handler de mensagens: %v\n", err)
		os.Exit(1)
	}

	waClient.AddEventHandler(msgHandler.HandleMessage)

	fmt.Println("Bot iniciado e escutando mensagens...")

	whatsapp.WaitForShutdown(waClient)

	fmt.Println("Bot desligado com sucesso.")
}
