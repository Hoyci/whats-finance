package whatsapp

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var (
	instance *WhatsappClient
	once     sync.Once
)

type WhatsappClient struct {
	Client *whatsmeow.Client
}

func GetInstance(dbPath string) *WhatsappClient {
	once.Do(func() {
		dbLog := waLog.Stdout("Database", "INFO", true)
		clientLog := waLog.Stdout("Client", "DEBUG", true)
		ctx := context.Background()

		fmt.Println(dbPath)

		if dbPath == "" {
			dbPath = "session.db"
		}

		container, err := sqlstore.New(ctx, "sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath), dbLog)
		if err != nil {
			log.Fatalf("Erro iniciando o banco de dados: %v", err)
		}

		deviceStore, err := container.GetFirstDevice(ctx)
		if err != nil {
			log.Fatalf("Falha ao pegar o primeiro device: %v", err)
		}
		client := whatsmeow.NewClient(deviceStore, clientLog)

		instance = &WhatsappClient{
			Client: client,
		}

		if instance.Client.Store.ID == nil {
			instance.ConnectionManager()
		} else {
			err := instance.Client.Connect()
			if err != nil {
				log.Panic("Erro conectando ao whatsapp", err)
			}
		}
	})

	return instance
}

func (c *WhatsappClient) WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Recebido sinal de desligamento. Fechando o cliente...")
	instance.Disconnect()
}
