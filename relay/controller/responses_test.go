package controller

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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
