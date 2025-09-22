package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/hoyci/whats-finance/internal/processor"
	googlesheets "github.com/hoyci/whats-finance/pkg/google-sheets"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type MessageHandler struct {
	TargetJID        types.JID
	ChatGPTProcessor *processor.ChatGPTProcessor
	SheetsService    *googlesheets.SheetsService
	WaClient         *whatsmeow.Client
	SheetName        string
}

func NewMessageHandler(
	waClient *whatsmeow.Client,
	targetJID string,
	chatGPTProcessor *processor.ChatGPTProcessor,
	sheetsService *googlesheets.SheetsService,
	sheetName string,
) (*MessageHandler, error) {
	jid := types.NewJID(targetJID, types.DefaultUserServer)

	return &MessageHandler{
		WaClient:         waClient,
		TargetJID:        jid,
		ChatGPTProcessor: chatGPTProcessor,
		SheetsService:    sheetsService,
		SheetName:        sheetName,
	}, nil
}

func (h *MessageHandler) HandleMessage(evt any) {
	msg, ok := evt.(*events.Message)
	fmt.Println("msg", msg)
	if !ok {
		return
	}

	if msg.Info.IsGroup {
		return
	}

	messageText := getMessageText(msg.Message)
	if messageText == "" {
		return
	}

	fmt.Printf("Mensagem recebida de %s: \"%s\"\n", msg.Info.Sender.User, messageText)
	go h.processAndRespond(h.TargetJID, messageText)
}

func (h *MessageHandler) processAndRespond(sender types.JID, message string) {
	sheetData, err := h.ChatGPTProcessor.ProcessMessage(message)
	if err != nil {
		fmt.Printf("Erro ao processar mensagem com ChatGPT: %v\n", err)
		h.sendWhatsAppMessage(sender, "Desculpe, não consegui processar sua solicitação agora. Tente novamente mais tarde.")
		return
	}

	err = h.appendWithRetry(sheetData)
	if err != nil {
		fmt.Printf("Erro ao inserir dados na planilha após múltiplas tentativas: %v\n", err)
		h.sendWhatsAppMessage(sender, "Sua transação foi processada, mas não foi possível registrá-la na planilha. Por favor, anote os detalhes e tente novamente mais tarde.")
		return
	}

	h.sendWhatsAppMessage(sender, sheetData.Retorno)
}

func (h *MessageHandler) sendWhatsAppMessage(to types.JID, text string) {
	_, err := h.WaClient.SendMessage(
		context.Background(),
		to,
		&waE2E.Message{
			Conversation: &text,
		},
	)
	if err != nil {
		fmt.Printf("Erro ao enviar mensagem para %s: %v\n", to, err)
	}
}

func getMessageText(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}

	if msg.GetConversation() != "" {
		return msg.GetConversation()
	}

	if extendedText := msg.GetExtendedTextMessage(); extendedText != nil {
		return extendedText.GetText()
	}

	if btnReply := msg.GetButtonsResponseMessage(); btnReply != nil {
		return btnReply.GetSelectedDisplayText()
	}

	return ""
}

func (h *MessageHandler) appendWithRetry(data *processor.SheetData) error {
	const maxRetries = 3
	var err error

	err = h.SheetsService.EnsureSheetExists(h.SheetName)
	if err != nil {
		return fmt.Errorf("falha ao garantir a existência da aba '%s': %w", h.SheetName, err)
	}

	for i := range maxRetries {
		err = h.SheetsService.AppendRow(h.SheetName, data)
		if err == nil {
			fmt.Printf("Dados inseridos com sucesso na planilha '%s'.\n", h.SheetName)
			return nil
		}

		fmt.Printf("Falha na tentativa %d/%d ao inserir dados: %v. Tentando novamente em %d segundos...\n", i+1, maxRetries, err, (i+1)*2)
		time.Sleep(time.Duration(i+1) * 2 * time.Second)
	}

	return fmt.Errorf("todas as %d tentativas de inserir dados na planilha falharam: %w", maxRetries, err)
}
