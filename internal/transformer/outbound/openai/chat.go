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

	// Log raw response for debugging reasoning issues
	log.Debugf("Raw response body: %s", string(body))

	var resp model.InternalLLMResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Post-process reasoning content for flexible format handling
	o.processReasoningContent(&resp, body)

	// Try to extract Gemini thinking from other possible locations
	o.extractGeminiThinking(&resp, body)

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
	o.processReasoningContent(&resp, eventData)

	return &resp, nil
}

// processReasoningContent handles reasoning content that might be in different formats
// Some providers return reasoning_content as an object {"text": "..."} or as a JSON string
func (o *ChatOutbound) processReasoningContent(resp *model.InternalLLMResponse, rawBody []byte) {
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
	// Log the raw content for debugging
	log.Debugf("Processing reasoning content: %s", content)

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
		if err := json.Unmarshal([]byte(content), &obj); err == nil {
			if obj.Text != "" {
				return &obj.Text
			}
			// If text is empty, return nil to indicate empty reasoning
			return nil
		}

		// Try parsing as {"content": "..."}
		var objContent struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(content), &objContent); err == nil {
			if objContent.Content != "" {
				return &objContent.Content
			}
			return nil
		}

		// Try parsing as {"reasoning": {"text": "..."}}
		var objNested struct {
			Reasoning struct {
				Text    string `json:"text"`
				Content string `json:"content"`
			} `json:"reasoning"`
		}
		if err := json.Unmarshal([]byte(content), &objNested); err == nil {
			if objNested.Reasoning.Text != "" {
				return &objNested.Reasoning.Text
			}
			if objNested.Reasoning.Content != "" {
				return &objNested.Reasoning.Content
			}
		}
	}

	// If it's an empty object or couldn't parse, return nil
	if content == "{}" || content == "{\"text\":\"\"}" || content == "" {
		return nil
	}

	// Return original if we couldn't process it
	return &content
}

// extractGeminiThinking attempts to find Gemini thinking content in various locations
func (o *ChatOutbound) extractGeminiThinking(resp *model.InternalLLMResponse, rawBody []byte) {
	// Some providers might include Gemini thinking in custom fields
	// Try to find it in various possible locations

	// Check each choice
	for i := range resp.Choices {
		if resp.Choices[i].Message != nil {
			// Check if usage indicates reasoning tokens were used
			if resp.Usage != nil && resp.Usage.CompletionTokensDetails != nil &&
				resp.Usage.CompletionTokensDetails.ReasoningTokens > 0 {

				// Check if we have reasoning_content but it's empty JSON
				if resp.Choices[i].Message.ReasoningContent != nil {
					content := *resp.Choices[i].Message.ReasoningContent
					if content == `{"text": ""}` || content == `{"text":""}` {
						log.Warnf("Gemini model used %d reasoning tokens but reasoning_content only contains empty JSON: %s",
							resp.Usage.CompletionTokensDetails.ReasoningTokens, content)

						// Try to find thinking in other fields
						var flexResp map[string]interface{}
						if err := json.Unmarshal(rawBody, &flexResp); err == nil {
							// Look for thinking in various possible locations
							thinking := o.findThinkingInMap(flexResp)
							if thinking != "" {
								log.Debugf("Found Gemini thinking content in alternative location")
								resp.Choices[i].Message.ReasoningContent = &thinking
							} else {
								// Check if provider sent reasoning in a separate field
								log.Errorf("Provider bug: Gemini used %d reasoning tokens but didn't include the actual thinking content. The provider is counting tokens but not sending the content.",
									resp.Usage.CompletionTokensDetails.ReasoningTokens)
							}
						}
					}
				} else {
					// No reasoning_content field at all
					log.Debugf("Gemini response has %d reasoning tokens but no reasoning_content field",
						resp.Usage.CompletionTokensDetails.ReasoningTokens)

					// Try to find thinking in other fields
					var flexResp map[string]interface{}
					if err := json.Unmarshal(rawBody, &flexResp); err == nil {
						thinking := o.findThinkingInMap(flexResp)
						if thinking != "" {
							log.Debugf("Found Gemini thinking content in alternative location")
							resp.Choices[i].Message.ReasoningContent = &thinking
						}
					}
				}
			}
		}
	}
}

// findThinkingInMap recursively searches for thinking content in response
func (o *ChatOutbound) findThinkingInMap(data map[string]interface{}) string {
	// Check common fields where providers might put thinking
	fields := []string{
		"thinking", "thought", "thoughts", "reasoning",
		"thinking_content", "thought_content", "internal_reasoning",
		"gemini_thinking", "gemini_thoughts", "model_thinking",
		"internal_thoughts", "reasoning_text", "thinking_text",
		"reasoning_process", "thought_process", "internal_monologue",
	}

	for _, field := range fields {
		if val, ok := data[field]; ok {
			if str, ok := val.(string); ok && str != "" {
				return str
			}
			// If it's a map, try to extract text from it
			if m, ok := val.(map[string]interface{}); ok {
				if text, ok := m["text"].(string); ok && text != "" {
					return text
				}
				if content, ok := m["content"].(string); ok && content != "" {
					return content
				}
			}
		}
	}

	// Check in choices if there are parts with thinking
	if choices, ok := data["choices"].([]interface{}); ok {
		for _, choice := range choices {
			if choiceMap, ok := choice.(map[string]interface{}); ok {
				// Check for parts array (Gemini-style)
				if message, ok := choiceMap["message"].(map[string]interface{}); ok {
					if parts, ok := message["parts"].([]interface{}); ok {
						for _, part := range parts {
							if partMap, ok := part.(map[string]interface{}); ok {
								// Check if this part is marked as thought
								if thought, ok := partMap["thought"].(bool); ok && thought {
									if text, ok := partMap["text"].(string); ok && text != "" {
										return text
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return ""
}
