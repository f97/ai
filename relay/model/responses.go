package model

// ResponsesRequest represents the OpenAI Responses API request format
// Minimal implementation supporting input and messages fields
type ResponsesRequest struct {
	Model    string    `json:"model" binding:"required"`
	Input    any       `json:"input,omitempty"`
	Messages []Message `json:"messages,omitempty"`
	Stream   bool      `json:"stream,omitempty"`
	// Additional fields that may be passed through
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
}

// ParseInput converts input field to messages format
// Returns nil if both input and messages are empty
func (r *ResponsesRequest) ParseInput() []Message {
	// If messages already provided, use them
	if len(r.Messages) > 0 {
		return r.Messages
	}

	// If no input, return empty
	if r.Input == nil {
		return nil
	}

	// Convert input to message
	var messages []Message

	switch v := r.Input.(type) {
	case string:
		messages = []Message{
			{
				Role:    "user",
				Content: v,
			},
		}
	case []any:
		// Handle array of strings/objects
		for _, item := range v {
			if str, ok := item.(string); ok {
				messages = append(messages, Message{
					Role:    "user",
					Content: str,
				})
			}
		}
	}

	return messages
}

// ResponsesOutputContent represents a content item in the output
type ResponsesOutputContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ResponsesOutputItem represents an output item (message) in the response
type ResponsesOutputItem struct {
	ID      string                    `json:"id"`
	Type    string                    `json:"type"`
	Role    string                    `json:"role"`
	Content []ResponsesOutputContent  `json:"content"`
}

// ResponsesResponse represents the OpenAI Responses API response format
type ResponsesResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created,omitempty"`
	Model   string                 `json:"model,omitempty"`
	Output  []ResponsesOutputItem  `json:"output"`
	Usage   *Usage                 `json:"usage,omitempty"`
}

// ResponsesStreamResponseOutputContent represents streaming output content
type ResponsesStreamResponseOutputContent struct {
	Type  string  `json:"type"`
	Delta *string `json:"delta,omitempty"`
	Text  *string `json:"text,omitempty"`
}

// ResponsesStreamResponseOutputItem represents a streaming output item
type ResponsesStreamResponseOutputItem struct {
	Index   int                                    `json:"index"`
	ID      *string                                `json:"id,omitempty"`
	Type    string                                 `json:"type"`
	Role    *string                                `json:"role,omitempty"`
	Content []ResponsesStreamResponseOutputContent `json:"content,omitempty"`
}

// ResponsesStreamResponse represents the streaming response format
type ResponsesStreamResponse struct {
	ID      string                              `json:"id"`
	Object  string                              `json:"object"`
	Created int64                               `json:"created,omitempty"`
	Model   string                              `json:"model,omitempty"`
	Output  []ResponsesStreamResponseOutputItem `json:"output,omitempty"`
	Usage   *Usage                              `json:"usage,omitempty"`
}
