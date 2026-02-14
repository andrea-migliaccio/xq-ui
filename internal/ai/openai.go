package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Provider represents an AI provider
type Provider interface {
	Name() string
	GenerateQuery(prompt, inputData, engineType string) (string, error)
	GenerateQueryWithContext(prompt, inputData, currentQuery, engineType string) (string, error)
	Configure(apiKey, model string) error
	IsConfigured() bool
}

// AIResult represents the result of AI query generation
type AIResult struct {
	Query       string
	Explanation string
	Success     bool
	Error       string
	Provider    string
	Duration    time.Duration
}

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey  string
	model   string
	baseURL string
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		model:   "gpt-4o-mini", // Default to cost-effective model
		baseURL: "https://api.openai.com/v1",
	}
}

func (p *OpenAIProvider) Name() string {
	return "OpenAI"
}

func (p *OpenAIProvider) Configure(apiKey, model string) error {
	p.apiKey = apiKey
	if model != "" {
		p.model = model
	}
	return nil
}

func (p *OpenAIProvider) IsConfigured() bool {
	return p.apiKey != ""
}

func (p *OpenAIProvider) GenerateQueryWithContext(prompt, inputData, currentQuery, engineType string) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("OpenAI provider not configured - missing API key")
	}
	
	// Create smart system message that includes current context
	var systemPrompt string
	if engineType == "jq" {
		systemPrompt = `You are a jq expert. Generate a jq query based on the user's request.

Current context:
- Engine: jq  
- User has input data available
- Current query (if any): ` + currentQuery + `

Instructions:
- Return ONLY the query, no explanation
- If user asks to "fix" or "correct", modify the current query
- If user describes desired output, create new query from scratch
- If user asks for transformation, analyze the request context

Examples:
- "get all names" → ".[] | .name"  
- "users older than 25" → ".users[] | select(.age > 25)"
- "fix my syntax" → (fix the current query syntax)
- "transform to show only names" → ".users[] | {name: .name}"`
	} else {
		systemPrompt = `You are a yq expert. Generate a yq query based on the user's request.

Current context:
- Engine: yq
- User has input data available  
- Current query (if any): ` + currentQuery + `

Instructions:
- Return ONLY the query, no explanation
- If user asks to "fix" or "correct", modify the current query
- If user describes desired output, create new query from scratch
- If user asks for transformation, analyze the request context

Examples:
- "get all names" → ".[] | .name"
- "users older than 25" → ".users[] | select(.age > 25)" 
- "fix my syntax" → (fix the current query syntax)
- "transform to show only names" → ".users[] | {\"name\": .name}"`
	}

	// Include input data context (truncate if too long)
	inputContext := inputData
	if len(inputContext) > 1000 {
		inputContext = inputContext[:1000] + "... (truncated)"
	}

	userPrompt := fmt.Sprintf(`Input data:
%s

Current query: %s

User request: %s

Generate the appropriate %s query:`, inputContext, currentQuery, prompt, engineType)

	// Create request payload
	payload := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens":    150,
		"temperature":   0.1,
		"stream":        false,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request with shorter timeout for UI responsiveness
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Make request with shorter timeout
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	query := response.Choices[0].Message.Content
	// Clean up the query (remove quotes, whitespace)
	query = cleanQuery(query)
	
	return query, nil
}

// GenerateQuery is the original method for backward compatibility
func (p *OpenAIProvider) GenerateQuery(prompt, inputData, engineType string) (string, error) {
	return p.GenerateQueryWithContext(prompt, inputData, "", engineType)
}

// cleanQuery removes common formatting artifacts from AI responses
func cleanQuery(query string) string {
	// Remove surrounding quotes if present
	if len(query) >= 2 && query[0] == '"' && query[len(query)-1] == '"' {
		query = query[1 : len(query)-1]
	}
	if len(query) >= 2 && query[0] == '`' && query[len(query)-1] == '`' {
		query = query[1 : len(query)-1]
	}
	
	// Remove jq/yq prefixes if AI added them
	if len(query) > 3 && (query[:3] == "jq " || query[:3] == "yq ") {
		query = query[3:]
	}
	
	return query
}

// LoadAPIKeyFromEnv loads API key from environment variable
func LoadOpenAIKeyFromEnv() string {
	return os.Getenv("OPENAI_API_KEY")
}