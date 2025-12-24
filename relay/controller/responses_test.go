package controller

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/stretchr/testify/assert"
)

func TestResponsesRequestParsing(t *testing.T) {
	// Test input as string
	t.Run("input as string", func(t *testing.T) {
		req := &model.ResponsesRequest{
			Model: "gpt-3.5-turbo",
			Input: "Hello, world!",
		}

		messages := req.ParseInput()
		assert.NotNil(t, messages)
		assert.Equal(t, 1, len(messages))
		assert.Equal(t, "user", messages[0].Role)
		assert.Equal(t, "Hello, world!", messages[0].Content)
	})

	// Test input as array
	t.Run("input as array", func(t *testing.T) {
		req := &model.ResponsesRequest{
			Model: "gpt-3.5-turbo",
			Input: []any{"Hello", "world"},
		}

		messages := req.ParseInput()
		assert.NotNil(t, messages)
		assert.Equal(t, 2, len(messages))
		assert.Equal(t, "user", messages[0].Role)
		assert.Equal(t, "Hello", messages[0].Content)
	})

	// Test messages provided directly
	t.Run("messages provided", func(t *testing.T) {
		req := &model.ResponsesRequest{
			Model: "gpt-3.5-turbo",
			Messages: []model.Message{
				{Role: "user", Content: "Test message"},
			},
		}

		messages := req.ParseInput()
		assert.NotNil(t, messages)
		assert.Equal(t, 1, len(messages))
		assert.Equal(t, "user", messages[0].Role)
	})

	// Test empty request
	t.Run("empty input", func(t *testing.T) {
		req := &model.ResponsesRequest{
			Model: "gpt-3.5-turbo",
		}

		messages := req.ParseInput()
		assert.Nil(t, messages)
	})
}

func TestResponsesRequestValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid request with input", func(t *testing.T) {
		reqBody := map[string]any{
			"model": "gpt-3.5-turbo",
			"input": "Hello, world!",
		}

		jsonData, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		var responsesReq model.ResponsesRequest
		err := json.NewDecoder(bytes.NewBuffer(jsonData)).Decode(&responsesReq)
		assert.NoError(t, err)
		assert.Equal(t, "gpt-3.5-turbo", responsesReq.Model)
		assert.NotNil(t, responsesReq.Input)
	})

	t.Run("valid request with messages", func(t *testing.T) {
		reqBody := map[string]any{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]string{
				{"role": "user", "content": "Hello"},
			},
		}

		jsonData, _ := json.Marshal(reqBody)
		var responsesReq model.ResponsesRequest
		err := json.NewDecoder(bytes.NewBuffer(jsonData)).Decode(&responsesReq)
		assert.NoError(t, err)
		assert.Equal(t, "gpt-3.5-turbo", responsesReq.Model)
		assert.NotNil(t, responsesReq.Messages)
		assert.Equal(t, 1, len(responsesReq.Messages))
	})

	t.Run("request with stream", func(t *testing.T) {
		reqBody := map[string]any{
			"model":  "gpt-3.5-turbo",
			"input":  "Hello",
			"stream": true,
		}

		jsonData, _ := json.Marshal(reqBody)
		var responsesReq model.ResponsesRequest
		err := json.NewDecoder(bytes.NewBuffer(jsonData)).Decode(&responsesReq)
		assert.NoError(t, err)
		assert.True(t, responsesReq.Stream)
	})
}

func TestChatCompletionToResponsesConversion(t *testing.T) {
	// Import OpenAI adapter for TextResponse type
	// This test verifies the conversion logic works correctly
	
	t.Run("convert valid chat completion", func(t *testing.T) {
		// Create a mock chat completion response
		chatResp := &openai.TextResponse{
			Id:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-3.5-turbo",
			Choices: []openai.TextResponseChoice{
				{
					Index: 0,
					Message: model.Message{
						Role:    "assistant",
						Content: "Hello! How can I help you today?",
					},
					FinishReason: "stop",
				},
			},
		}
		chatResp.Usage = model.Usage{
			PromptTokens:     10,
			CompletionTokens: 8,
			TotalTokens:      18,
		}

		// Convert to Responses format
		responsesResp := convertChatCompletionToResponses(chatResp)

		// Verify conversion
		assert.NotNil(t, responsesResp)
		assert.Equal(t, "chatcmpl-123", responsesResp.ID)
		assert.Equal(t, "response", responsesResp.Object)
		assert.Equal(t, int64(1234567890), responsesResp.Created)
		assert.Equal(t, "gpt-3.5-turbo", responsesResp.Model)
		assert.NotNil(t, responsesResp.Usage)
		assert.Equal(t, 10, responsesResp.Usage.PromptTokens)
		assert.Equal(t, 8, responsesResp.Usage.CompletionTokens)
		assert.Equal(t, 18, responsesResp.Usage.TotalTokens)

		// Verify output structure
		assert.Equal(t, 1, len(responsesResp.Output))
		assert.Equal(t, "message", responsesResp.Output[0].Type)
		assert.Equal(t, "assistant", responsesResp.Output[0].Role)
		assert.Equal(t, 1, len(responsesResp.Output[0].Content))
		assert.Equal(t, "output_text", responsesResp.Output[0].Content[0].Type)
		assert.Equal(t, "Hello! How can I help you today?", responsesResp.Output[0].Content[0].Text)
	})

	t.Run("convert empty chat completion", func(t *testing.T) {
		// Create a chat completion with no choices (should not happen in practice)
		chatResp := &openai.TextResponse{
			Id:      "chatcmpl-456",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-3.5-turbo",
			Choices: []openai.TextResponseChoice{},
		}

		// Convert to Responses format
		responsesResp := convertChatCompletionToResponses(chatResp)

		// Should still create a valid response structure
		assert.NotNil(t, responsesResp)
		assert.Equal(t, "chatcmpl-456", responsesResp.ID)
		assert.Equal(t, "response", responsesResp.Object)
		assert.Equal(t, 0, len(responsesResp.Output))
	})
}
