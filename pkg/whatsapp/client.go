package whatsapp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
)

func InitializeClient() (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "INFO", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:session.db?_foreign_keys=on", dbLog)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar o container de armazenamento: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}

		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("Escaneie este QR Code no seu WhatsApp:")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Printf("Login status: %s\n", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	return client, nil
}

func WaitForShutdown(client *whatsmeow.Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Recebido sinal de desligamento. Fechando o cliente...")
	client.Disconnect()
}
