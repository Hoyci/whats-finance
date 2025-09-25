package handler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hoyci/whats-finance/internal/processor"
	googlesheets "github.com/hoyci/whats-finance/pkg/google-sheets"
	"github.com/hoyci/whats-finance/pkg/whatsapp"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type BotHandler struct {
	whatsappClient   *whatsapp.WhatsappClient
	chatGPTProcessor *processor.ChatGPTProcessor
	sheetsService    *googlesheets.SheetsService
	sheetName        string
	targetJID        types.JID
}

func NewBotHandler(
	whatsappClient *whatsapp.WhatsappClient,
	chatGPTProcessor *processor.ChatGPTProcessor,
	sheetsService *googlesheets.SheetsService,
	sheetName string,
	targetJID string,
) *BotHandler {
	jid := types.NewJID(targetJID, types.DefaultUserServer)

	return &BotHandler{
		whatsappClient:   whatsappClient,
		chatGPTProcessor: chatGPTProcessor,
		sheetsService:    sheetsService,
		sheetName:        sheetName,
		targetJID:        jid,
	}
}

func (h *BotHandler) InitializeBot() {
	h.whatsappClient.Client.AddEventHandler(h.eventHandler)
	h.handleShutdown()
}

func (h *BotHandler) processMessageAndRespond(event *events.Message) {
	from := event.Info.Sender
	msg := h.handleIncomingMessage(event)
	sheetData, err := h.chatGPTProcessor.ProcessMessage(msg)
	if err != nil {
		fmt.Printf("Erro ao processar mensagem com ChatGPT: %v\n", err)
		h.handleOutgoingMessage(from, "Desculpe, não consegui processar sua solicitação agora. Tente novamente mais tarde.")
		return
	}

	err = h.appendWithRetry(sheetData)
	if err != nil {
		fmt.Printf("Erro ao inserir dados na planilha após múltiplas tentativas: %v\n", err)
		h.handleOutgoingMessage(from, "Sua transação foi processada, mas não foi possível registrá-la na planilha. Por favor, anote os detalhes e tente novamente mais tarde.")
		return
	}

	h.handleOutgoingMessage(from, sheetData.Retorno)
}

func (h *BotHandler) appendWithRetry(data *processor.SheetData) error {
	const maxRetries = 3
	var err error

	err = h.sheetsService.EnsureSheetExists(h.sheetName)
	if err != nil {
		return fmt.Errorf("falha ao garantir a existência da aba '%s': %w", h.sheetName, err)
	}

	for i := range maxRetries {
		err = h.sheetsService.AppendRow(h.sheetName, data)
		if err == nil {
			fmt.Printf("Dados inseridos com sucesso na planilha '%s'.\n", h.sheetName)
			return nil
		}

		fmt.Printf("Falha na tentativa %d/%d ao inserir dados: %v. Tentando novamente em %d segundos...\n", i+1, maxRetries, err, (i+1)*2)
		time.Sleep(time.Duration(i+1) * 2 * time.Second)
	}

	return fmt.Errorf("todas as %d tentativas de inserir dados na planilha falharam: %w", maxRetries, err)
}

func (c *BotHandler) handleIncomingMessage(event *events.Message) string {
	if event.Info.IsFromMe || event.Info.IsGroup {
		return ""
	}

	msg := event.Message.GetConversation()
	if msg == "" {
		if mediaMsg := event.Message.GetExtendedTextMessage().GetText(); mediaMsg != "" {
			msg = mediaMsg
		} else {
			return ""
		}
	}

	fmt.Printf("Mensagem recebida de %s: \"%s\"\n", event.Info.Sender.User, msg)
	return msg
}

func (h *BotHandler) handleOutgoingMessage(to types.JID, msg string) {
	ctx := context.Background()
	res, err := h.whatsappClient.Client.SendMessage(
		ctx,
		to,
		&waE2E.Message{
			Conversation: &msg,
		},
	)
	if err != nil {
		fmt.Printf("Erro ao enviar mensagem para %s: %v\n", to, err)
	}

	fmt.Printf("Messagem enviada com sucesso! %+v", res)
}

func (h *BotHandler) eventHandler(evt any) {
	switch e := evt.(type) {
	case *events.Connected:
		log.Println("Connected successfully")
	case *events.Disconnected:
		log.Println("Disconnected, trying to reconnect")
		go h.whatsappClient.ConnectionManager()
	case *events.Message:
		h.processMessageAndRespond(e)
	}
}

func (h *BotHandler) handleShutdown() {
	h.whatsappClient.WaitForShutdown()
	fmt.Println("Bot desligado com sucesso.")
}
