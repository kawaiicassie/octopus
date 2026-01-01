package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/bestruirui/octopus/internal/transformer/model"
	"github.com/bestruirui/octopus/internal/utils/log"
)

type ChatOutbound struct{}

func (o *ChatOutbound) TransformRequest(ctx context.Context, request *model.InternalLLMRequest, baseUrl, key string) (*http.Request, error) {
	request.ClearHelpFields()

	// Convert developer role to system role for compatibility
	for i := range request.Messages {
		if request.Messages[i].Role == "developer" {
			request.Messages[i].Role = "system"
		}
	}

	if request.Stream != nil && *request.Stream {
		if request.StreamOptions == nil {
			request.StreamOptions = &model.StreamOptions{IncludeUsage: true}
		} else if !request.StreamOptions.IncludeUsage {
			request.StreamOptions.IncludeUsage = true
		}
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	parsedUrl, err := url.Parse(strings.TrimSuffix(baseUrl, "/"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %w", err)
	}
	parsedUrl.Path = parsedUrl.Path + "/chat/completions"
	req.URL = parsedUrl
	req.Method = http.MethodPost
	return req, nil
}

func (o *ChatOutbound) TransformResponse(ctx context.Context, response *http.Response) (*model.InternalLLMResponse, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("response body is empty")
	}

	var resp model.InternalLLMResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Post-process reasoning content for flexible format handling
	o.processReasoningContent(&resp)

	return &resp, nil
}

func (o *ChatOutbound) TransformStream(ctx context.Context, eventData []byte) (*model.InternalLLMResponse, error) {
	if bytes.HasPrefix(eventData, []byte("[DONE]")) {
		return &model.InternalLLMResponse{
			Object: "[DONE]",
		}, nil
	}

	var errCheck struct {
		Error *model.ErrorDetail `json:"error"`
	}
	if err := json.Unmarshal(eventData, &errCheck); err == nil && errCheck.Error != nil {
		return nil, &model.ResponseError{
			Detail: *errCheck.Error,
		}
	}

	var resp model.InternalLLMResponse
	if err := json.Unmarshal(eventData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stream chunk: %w", err)
	}

	// Post-process reasoning content for flexible format handling in streaming
	o.processReasoningContent(&resp)

	return &resp, nil
}

// processReasoningContent handles reasoning content that might be in different formats
// Some providers return reasoning_content as an object {"text": "..."} or as a JSON string
func (o *ChatOutbound) processReasoningContent(resp *model.InternalLLMResponse) {
	// Process each choice
	for i := range resp.Choices {
		// Check message reasoning content
		if resp.Choices[i].Message != nil && resp.Choices[i].Message.ReasoningContent != nil {
			processed := o.extractReasoningText(*resp.Choices[i].Message.ReasoningContent)
			if processed != nil {
				resp.Choices[i].Message.ReasoningContent = processed
			}
			// Keep the original content if we couldn't process it
			// This preserves the field for providers that might send content differently
		}

		// Check delta reasoning content (for streaming)
		if resp.Choices[i].Delta != nil && resp.Choices[i].Delta.ReasoningContent != nil {
			processed := o.extractReasoningText(*resp.Choices[i].Delta.ReasoningContent)
			if processed != nil {
				resp.Choices[i].Delta.ReasoningContent = processed
			}
			// Keep the original content if we couldn't process it
		}
	}
}

// extractReasoningText extracts the actual reasoning text from various formats
func (o *ChatOutbound) extractReasoningText(content string) *string {
	// If it's already a non-empty normal string, return as-is
	if content != "" && !strings.HasPrefix(content, "{") && !strings.HasPrefix(content, "[") {
		return &content
	}

	// Try to parse as JSON object
	if strings.HasPrefix(content, "{") {
		// Try parsing as {"text": "..."}
		var obj struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal([]byte(content), &obj); err == nil && obj.Text != "" {
			return &obj.Text
		}

		// Try parsing as {"content": "..."}
		var objContent struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(content), &objContent); err == nil && objContent.Content != "" {
			return &objContent.Content
		}
	}

	// Return original if we couldn't process it (keep the field as-is)
	// This preserves whatever format the provider sent
	if content != "" {
		return &content
	}

	return nil
}

