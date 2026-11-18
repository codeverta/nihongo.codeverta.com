package aichat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ChatMessage is one OpenAI-compatible chat message.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Messages    []ChatMessage
	MaxTokens   int
	Temperature float64
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	CacheHitTokens   int `json:"prompt_cache_hit_tokens"`
	CacheMissTokens  int `json:"prompt_cache_miss_tokens"`
}

type Completion struct {
	Text         string
	FinishReason string
	Usage        TokenUsage
}

// Completer streams a model answer, calling onDelta for each text fragment.
type Completer interface {
	Stream(ctx context.Context, req CompletionRequest, onDelta func(string) error) (Completion, error)
}

// DeepSeekClient calls DeepSeek's OpenAI-compatible chat completions API.
type DeepSeekClient struct {
	APIKey  string
	BaseURL string
	Model   string
	HTTP    *http.Client
}

func NewDeepSeekClient(apiKey, baseURL, model string) *DeepSeekClient {
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	if model == "" {
		model = "deepseek-chat"
	}
	// No overall timeout: streams may run long. Callers bound them via ctx.
	transport := &http.Transport{ResponseHeaderTimeout: 30 * time.Second, IdleConnTimeout: 90 * time.Second}
	return &DeepSeekClient{APIKey: apiKey, BaseURL: strings.TrimRight(baseURL, "/"), Model: model, HTTP: &http.Client{Transport: transport}}
}

// UpstreamError is a non-2xx answer from DeepSeek. Its body is never shown to
// end users; it is only logged.
type UpstreamError struct {
	Status int
	Body   string
}

func (e *UpstreamError) Error() string { return fmt.Sprintf("deepseek: HTTP %d: %s", e.Status, e.Body) }

func (c *DeepSeekClient) Stream(ctx context.Context, req CompletionRequest, onDelta func(string) error) (Completion, error) {
	payload := map[string]any{
		"model":          c.Model,
		"messages":       req.Messages,
		"stream":         true,
		"max_tokens":     req.MaxTokens,
		"temperature":    req.Temperature,
		"stream_options": map[string]bool{"include_usage": true},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Completion{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Completion{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return Completion{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return Completion{}, &UpstreamError{Status: resp.StatusCode, Body: string(snippet)}
	}
	return readSSE(resp.Body, onDelta)
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *TokenUsage `json:"usage"`
}

func readSSE(r io.Reader, onDelta func(string) error) (Completion, error) {
	var out Completion
	var text strings.Builder
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	done := false
	for scanner.Scan() {
		line := scanner.Text()
		// Blank lines separate events; ":" lines are keep-alive comments.
		if line == "" || strings.HasPrefix(line, ":") || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			done = true
			break
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return out, fmt.Errorf("deepseek: decode stream chunk: %w", err)
		}
		if chunk.Usage != nil {
			out.Usage = *chunk.Usage
		}
		for _, choice := range chunk.Choices {
			if choice.FinishReason != nil {
				out.FinishReason = *choice.FinishReason
			}
			if choice.Delta.Content == "" {
				continue
			}
			text.WriteString(choice.Delta.Content)
			if err := onDelta(choice.Delta.Content); err != nil {
				return out, err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return out, err
	}
	out.Text = text.String()
	if !done && out.FinishReason == "" {
		return out, errors.New("deepseek: stream ended unexpectedly")
	}
	return out, nil
}
