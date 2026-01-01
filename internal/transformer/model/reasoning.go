package model

import (
	"encoding/json"
	"strings"
)

// FlexibleReasoningContent handles reasoning content that can be either a string or an object
type FlexibleReasoningContent struct {
	Value *string
}

// UnmarshalJSON handles both string and object formats for reasoning content
func (f *FlexibleReasoningContent) UnmarshalJSON(data []byte) error {
	// First try to unmarshal as a string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		f.Value = &str
		return nil
	}

	// If that fails, try as an object with "text" field
	var obj struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		// Only set if text is not empty
		if obj.Text != "" {
			f.Value = &obj.Text
		}
		return nil
	}

	// Try as an object with "content" field (some providers might use this)
	var objContent struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(data, &objContent); err == nil {
		if objContent.Content != "" {
			f.Value = &objContent.Content
		}
		return nil
	}

	// Try as an object with nested structure
	var nestedObj struct {
		Reasoning struct {
			Text    string `json:"text"`
			Content string `json:"content"`
		} `json:"reasoning"`
	}
	if err := json.Unmarshal(data, &nestedObj); err == nil {
		if nestedObj.Reasoning.Text != "" {
			f.Value = &nestedObj.Reasoning.Text
		} else if nestedObj.Reasoning.Content != "" {
			f.Value = &nestedObj.Reasoning.Content
		}
		return nil
	}

	// If all parsing attempts fail, try to get raw string representation
	// Remove quotes and clean up
	rawStr := string(data)
	rawStr = strings.Trim(rawStr, `"`)
	if rawStr != "" && rawStr != "null" && rawStr != "{}" {
		f.Value = &rawStr
	}

	return nil
}

// MarshalJSON marshals the reasoning content back to JSON
func (f FlexibleReasoningContent) MarshalJSON() ([]byte, error) {
	if f.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*f.Value)
}

// String returns the string value or empty string if nil
func (f *FlexibleReasoningContent) String() string {
	if f.Value == nil {
		return ""
	}
	return *f.Value
}

// MessageWithFlexibleReasoning extends Message to handle flexible reasoning content
type MessageWithFlexibleReasoning struct {
	Message
	FlexibleReasoningContent *FlexibleReasoningContent `json:"reasoning_content,omitempty"`
}

// PostProcessReasoning converts FlexibleReasoningContent back to standard ReasoningContent
func (m *MessageWithFlexibleReasoning) PostProcessReasoning() {
	if m.FlexibleReasoningContent != nil && m.FlexibleReasoningContent.Value != nil {
		m.ReasoningContent = m.FlexibleReasoningContent.Value
		// Clear the flexible field after processing
		m.FlexibleReasoningContent = nil
	}
}