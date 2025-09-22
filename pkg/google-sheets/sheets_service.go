package googlesheets

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/hoyci/whats-finance/internal/processor"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetsService struct {
	client        *http.Client
	sheetsService *sheets.Service
	spreadsheetID string
}

func NewSheetsService(credentialsFile, spreadsheetID string) (*SheetsService, error) {
	ctx := context.Background()

	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler o arquivo de credenciais: %w", err)
	}

	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("falha ao configurar o cliente JWT: %w", err)
	}

	client := config.Client(ctx)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar o cliente do Google Sheets: %w", err)
	}

	return &SheetsService{
		client:        client,
		sheetsService: srv,
		spreadsheetID: spreadsheetID,
	}, nil
}

func (s *SheetsService) AppendRow(sheetName string, data *processor.SheetData) error {
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []any{data.Data, data.Tipo, data.Valor, data.Descricao})

	_, err := s.sheetsService.Spreadsheets.Values.Append(s.spreadsheetID, fmt.Sprintf("%s!A:D", sheetName), &vr).
		ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("falha ao adicionar a linha à planilha: %w", err)
	}

	fmt.Printf("Dados adicionados com sucesso à planilha: %s\n", sheetName)
	return nil
}

func (s *SheetsService) EnsureSheetExists(sheetName string) error {
	resp, err := s.sheetsService.Spreadsheets.Get(s.spreadsheetID).Fields("sheets.properties.title").Do()
	if err != nil {
		return fmt.Errorf("falha ao obter a planilha: %w", err)
	}

	for _, sheet := range resp.Sheets {
		if sheet.Properties.Title == sheetName {
			return nil // A aba já existe
		}
	}

	requests := []*sheets.Request{
		{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: sheetName,
				},
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	_, err = s.sheetsService.Spreadsheets.BatchUpdate(s.spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("falha ao criar a aba %s: %w", sheetName, err)
	}

	fmt.Printf("Aba '%s' criada na planilha.\n", sheetName)
	return nil
}
