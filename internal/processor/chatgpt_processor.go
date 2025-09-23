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
	Retorno   string  `json:"retorno"`
	Categoria string  `json:"categoria"`
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
	return &ChatGPTProcessor{
		apiKey: apiKey,
	}
}

func (p *ChatGPTProcessor) ProcessMessage(message string) (*SheetData, error) {
	prompt := fmt.Sprintf(`
Você é um assistente financeiro divertido e organizado.
Analise a mensagem do usuário (que pode ser uma receita ou despesa) e **retorne sempre um JSON válido, bem formatado, sem texto adicional**.

O JSON deve conter os seguintes campos obrigatórios:
* **tipo**: "receita" ou "despesa"
* **valor**: número em reais (sem "R\$")
* **descricao**: resumo curto do gasto ou receita
* **data**: formato "dd/mm/aaaa" (se não especificado, usar a data de hoje)
* **categoria**: uma entre \["alimentação", "transporte", "moradia", "lazer", "saúde", "outros"]
* **retorno**: uma frase divertida e bem jovial, entre 20 e 50 palavras, com emojis, mencionando o tipo de movimentação, o valor, a descrição e a categoria.

O JSON sempre deve estar formatado e pronto para ser unmarshallized! Em hipótese alguma ele deve estar fora do formato padrão de JSON.
### Exemplos:

Mensagem: "Paguei o almoço hoje, foi 25 reais."
JSON:
{"tipo":"despesa","valor":25,"descricao":"almoço","data":"%s","categoria":"alimentação","retorno":"📉 Você torrou R$25,00 em Alimentação 🍽️. Um rango top que deixou o bolso mais leve 💸, mas valeu a pena pra matar a fome e curtir o momento 😋🔥"}

Mensagem: "Recebi o pagamento do cliente, 500 reais."
JSON:
{"tipo":"receita","valor":500,"descricao":"pagamento do cliente","data":"%s","categoria":"outros","retorno":"📈 R$500,00 de Receita chegaram no seu caixa 💼🚀. É aquele up na conta que anima o dia, enche o bolso e dá até vontade de comemorar com um rolê 🎉💰"}

Mensagem: "Comprei um livro por 50 na semana passada."
JSON:
{"tipo":"despesa","valor":50,"descricao":"compra de livro","data":"%s","categoria":"lazer","retorno":"📉 Você gastou R$50,00 em Lazer 📚. Investiu numa boa leitura que vai abrir a mente 🤓✨. O bolso chora um pouquinho 💸, mas o cérebro agradece muito 📖🔥"}

Mensagem do usuário: "%s"
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
