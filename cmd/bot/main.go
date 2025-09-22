package main

import (
	"fmt"
	"os"

	"github.com/hoyci/whats-finance/internal/config"
	"github.com/hoyci/whats-finance/internal/handler"
	"github.com/hoyci/whats-finance/internal/processor"
	"github.com/hoyci/whats-finance/pkg/whatsapp"
)

func main() {
	cfg := config.GetConfig()
	if cfg == nil {
		fmt.Println("Falha ao carregar as configurações. Verifique o arquivo .env.")
		os.Exit(1)
	}

	waClient, err := whatsapp.InitializeClient()
	if err != nil {
		fmt.Printf("Falha ao inicializar o cliente do WhatsApp: %v\n", err)
		os.Exit(1)
	}

	chatGPTProcessor := processor.NewChatGPTProcessor(cfg.ChatGPTApiKey)

	msgHandler, err := handler.NewMessageHandler(waClient, cfg.PhoneNumber, chatGPTProcessor)
	if err != nil {
		fmt.Printf("Falha ao criar o handler de mensagens: %v\n", err)
		os.Exit(1)
	}

	waClient.AddEventHandler(msgHandler.HandleMessage)

	fmt.Println("Bot iniciado e escutando mensagens...")

	whatsapp.WaitForShutdown(waClient)

	fmt.Println("Bot desligado com sucesso.")
}
