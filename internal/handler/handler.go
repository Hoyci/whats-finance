package handler

import (
	"context"
	"fmt"

	"github.com/hoyci/whats-finance/internal/processor"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// Handler de Mensagens
type MessageHandler struct {
	TargetJID        types.JID
	ChatGPTProcessor *processor.ChatGPTProcessor
	WaClient         *whatsmeow.Client
}

// NewMessageHandler cria e retorna um novo handler
func NewMessageHandler(waClient *whatsmeow.Client, targetJID string, chatGPTProcessor *processor.ChatGPTProcessor) (*MessageHandler, error) {
	// jid, err := types.ParseJID(fmt.Sprintf(`%s@s.whatsapp.net`, targetJID))
	jid := types.NewJID(targetJID, types.DefaultUserServer)
	fmt.Println("JID", jid.String())
	// if err != nil {
	// 	return nil, fmt.Errorf("invalid target JID: %w", err)
	// }

	return &MessageHandler{
		WaClient:         waClient,
		TargetJID:        jid,
		ChatGPTProcessor: chatGPTProcessor,
	}, nil
}

// HandleMessage é a função principal que processa os eventos recebidos
func (h *MessageHandler) HandleMessage(evt any) {
	// A mensagem vem como um evento do tipo *event.MessageInfo
	msg, ok := evt.(*events.Message)
	fmt.Println("msg", msg)
	if !ok {
		return // Ignora eventos que não são de mensagem
	}

	// Filtra a mensagem para garantir que ela venha do contato desejado
	if msg.Info.IsGroup {
		fmt.Println("cai aqui")
		return
	}

	fmt.Println("cheguei aqui")

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

	confirmationMessage := fmt.Sprintf("Receita/despesa registrada: %s de R$%.2f (Descrição: %s).", sheetData.Tipo, sheetData.Valor, sheetData.Descricao)
	h.sendWhatsAppMessage(sender, confirmationMessage)
}

func (h *MessageHandler) sendWhatsAppMessage(to types.JID, text string) {
	// Use a nova struct waProto.Message
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

	// Check for a simple text conversation
	if msg.GetConversation() != "" {
		return msg.GetConversation()
	}

	// Check for an extended text message
	if extendedText := msg.GetExtendedTextMessage(); extendedText != nil {
		return extendedText.GetText()
	}

	// Check for a message with a button reply
	if btnReply := msg.GetButtonsResponseMessage(); btnReply != nil {
		return btnReply.GetSelectedDisplayText()
	}

	return ""
}
