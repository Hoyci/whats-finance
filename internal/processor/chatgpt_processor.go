package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SheetData struct {
	Tipo      string  `json:"tipo"`
	Valor     float64 `json:"valor"`
	Descricao string  `json:"descricao"`
	Data      string  `json:"data"`
}

type chatGPTRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type chatGPTResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type ChatGPTProcessor struct {
	apiKey string
}

func NewChatGPTProcessor(apiKey string) *ChatGPTProcessor {
	fmt.Println("apiKey", apiKey)
	return &ChatGPTProcessor{
		apiKey: apiKey,
	}
}

func (p *ChatGPTProcessor) ProcessMessage(message string) (*SheetData, error) {
	prompt := fmt.Sprintf(`
Você é um assistente financeiro. Analise a seguinte mensagem do usuário, que se refere a uma receita ou despesa.
Retorne um JSON com os seguintes campos: 'tipo' (receita/despesa), 'valor', 'descricao' e 'data'.
A data deve ser no formato 'dd/mm/aaaa'. Se a data não for especificada, use a data de hoje.

Exemplos:
- Mensagem: "Paguei o almoço hoje, foi 25 reais."
  JSON: {"tipo":"despesa","valor":25,"descricao":"almoço","data":"%s"}

- Mensagem: "Recebi o pagamento do cliente, 500 reais."
  JSON: {"tipo":"receita","valor":500,"descricao":"pagamento do cliente","data":"%s"}

- Mensagem: "Comprei um livro por 50 na semana passada."
  JSON: {"tipo":"despesa","valor":50,"descricao":"compra de livro","data":"%s"}

Mensagem do usuário: "%s"

Lembre-se de retornar apenas o JSON, sem texto adicional.
`, time.Now().Format("02/01/2006"), time.Now().Format("02/01/2006"), time.Now().AddDate(0, 0, -7).Format("02/01/2006"), message)

	requestBody := chatGPTRequest{
		Model: "gpt-3.5-turbo",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "system",
				Content: "Você é um assistente útil que retorna JSON.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar a requisição: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar a requisição HTTP: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao chamar a API do OpenAI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro na API do OpenAI: status %d, corpo: %s", resp.StatusCode, string(body))
	}

	var chatResponse chatGPTResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		return nil, fmt.Errorf("falha ao decodificar a resposta JSON: %w", err)
	}

	if len(chatResponse.Choices) == 0 {
		return nil, fmt.Errorf("a API não retornou uma resposta válida")
	}

	rawJSON := chatResponse.Choices[0].Message.Content
	var sheetData SheetData
	if err := json.Unmarshal([]byte(rawJSON), &sheetData); err != nil {
		return nil, fmt.Errorf("falha ao decodificar o JSON da resposta do ChatGPT: %w", err)
	}

	return &sheetData, nil
}
