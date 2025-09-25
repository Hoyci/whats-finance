package whatsapp

import (
	"context"
	"fmt"
	"log"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
)

func (c *WhatsappClient) ConnectionManager() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	qrChann, err := c.Client.GetQRChannel(ctx)
	if err != nil {
		log.Printf("erro acessando o qrcode: %v", err)
		return
	}

	err = c.Client.Connect()
	if err != nil {
		log.Printf("Erro na conexão inicial: %v", err)
		return
	}

	for item := range qrChann {
		switch item.Event {
		case whatsmeow.QRChannelEventCode:
			c.generateAndDisplayQR(item.Code)
		case whatsmeow.QRChannelSuccess.Event:
			log.Println("Logado com sucesso")
			cancel()
		case whatsmeow.QRChannelTimeout.Event:
			log.Println("QRCode expirado")
			c.ConnectionManager()
		case whatsmeow.QRChannelEventError:
			log.Printf("Erro de autenticação: %v", item.Error)
			c.ConnectionManager()
		}
	}
}

func (c *WhatsappClient) Disconnect() func() {
	return c.Client.Disconnect
}

func (c *WhatsappClient) generateAndDisplayQR(code string) {
	qr, err := qrcode.New(code, qrcode.Low)
	if err != nil {
		log.Printf("Erro ao tentar gerar o qrcode: %v", err)
		return
	}
	qr.DisableBorder = true

	fmt.Println("\n=== QR CODE ===")
	fmt.Println(qr.ToSmallString(false))
	fmt.Println("====================")
}
