# Channel Types Guide for Octopus

## Important: Choose the Right Channel Type

The Channel Type determines how Octopus transforms requests and responses. **Choose based on the API format, NOT the model provider!**

## Channel Type Selection Guide

### Use "OpenAI Chat" when:
- The endpoint follows OpenAI API format (`/v1/chat/completions`)
- Response includes standard OpenAI fields like `choices`, `message`, `reasoning_content`
- Examples:
  - OpenAI API
  - Azure OpenAI
  - OpenRouter (all models including Gemini, Claude)
  - Most proxy services
  - **Gemini models via OpenAI-compatible endpoints**
  - DeepSeek API
  - Together AI
  - Groq

### Use "OpenAI Responses" when:
- Using OpenAI Responses API (`/v1/responses`)
- Need advanced features like prediction, audio
- Limited provider support

### Use "Anthropic" when:
- Connecting directly to Anthropic API (`api.anthropic.com`)
- Using native Anthropic format (`/v1/messages`)
- Need Anthropic-specific features like thinking blocks
- **Note:** Thinking content becomes SSE events, may not work with OpenAI clients

### Use "Gemini" when:
- Connecting directly to Google AI API (`generativelanguage.googleapis.com`)
- Using native Gemini format
- **NOT for Gemini models via OpenAI-compatible endpoints!**

### Use "Volcengine" when:
- Using Volcengine/Doubao APIs
- China-specific endpoint

## Common Mistakes

### ❌ WRONG: Setting Channel Type based on model name
```yaml
Channel:
  Name: "Gemini on OpenRouter"
  Type: Gemini  # WRONG! OpenRouter uses OpenAI format
  Base URL: https://openrouter.ai/api/v1
  Model: google/gemini-2.0-flash
```

### ✅ CORRECT: Setting Channel Type based on API format
```yaml
Channel:
  Name: "Gemini on OpenRouter"
  Type: OpenAI Chat  # CORRECT! OpenRouter uses OpenAI format
  Base URL: https://openrouter.ai/api/v1
  Model: google/gemini-2.0-flash
```

## Reasoning/Thinking Support by Channel Type

| Channel Type | Reasoning Support | Format |
|--------------|-------------------|---------|
| **OpenAI Chat** | ✅ Full | `reasoning_content` field in response |
| **OpenAI Responses** | ✅ Full | `reasoning_summary_text` in response |
| **Anthropic** | ⚠️ Partial | Thinking blocks as SSE events (complex) |
| **Gemini** | ❌ Limited | Only for direct Gemini API |
| **Volcengine** | ✅ Model-specific | Only certain models support |

## How to Test

1. Create a channel with your endpoint
2. Send a test request with `reasoning_effort: "high"`
3. Check if `reasoning_content` appears in response
4. If not working, try changing Channel Type to "OpenAI Chat"

## Rule of Thumb

**When in doubt, use "OpenAI Chat"** - it's the most compatible format supported by most providers and proxies.