package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

// RelayResponsesHelper handles the /v1/responses endpoint
// by converting to/from chat completions format internally
func RelayResponsesHelper(c *gin.Context) *model.ErrorWithStatusCode {
	ctx := c.Request.Context()

	// Parse responses request
	responsesReq := &model.ResponsesRequest{}
	err := common.UnmarshalBodyReusable(c, responsesReq)
	if err != nil {
		logger.Errorf(ctx, "failed to parse responses request: %s", err.Error())
		return openai.ErrorWrapper(err, "invalid_request_error", http.StatusBadRequest)
	}

	// Validate request
	if responsesReq.Model == "" {
		return openai.ErrorWrapper(fmt.Errorf("model is required"), "invalid_request_error", http.StatusBadRequest)
	}

	// Convert to messages format
	messages := responsesReq.ParseInput()
	if len(messages) == 0 {
		return openai.ErrorWrapper(fmt.Errorf("either input or messages must be provided"), "invalid_request_error", http.StatusBadRequest)
	}

	// Create a ChatCompletion request from Responses request
	chatReq := &model.GeneralOpenAIRequest{
		Model:       responsesReq.Model,
		Messages:    messages,
		Stream:      responsesReq.Stream,
		MaxTokens:   responsesReq.MaxTokens,
		Temperature: responsesReq.Temperature,
		TopP:        responsesReq.TopP,
	}

	// Marshal back to JSON and set as request body for downstream processing
	reqBody, err := json.Marshal(chatReq)
	if err != nil {
		logger.Errorf(ctx, "failed to marshal chat request: %s", err.Error())
		return openai.ErrorWrapper(err, "internal_error", http.StatusInternalServerError)
	}

	// Replace request body with chat completions format
	c.Request.Body = io.NopCloser(strings.NewReader(string(reqBody)))
	c.Request.ContentLength = int64(len(reqBody))

	// Store original stream setting for response conversion
	c.Set("responses_mode", true)
	c.Set("responses_stream", responsesReq.Stream)

	// Call the existing text relay helper which handles all the pipeline
	// We need to intercept the response to convert it back to Responses format
	if responsesReq.Stream {
		return relayResponsesStream(c, chatReq)
	} else {
		return relayResponsesNonStream(c, chatReq)
	}
}

// relayResponsesNonStream handles non-streaming responses
func relayResponsesNonStream(c *gin.Context, chatReq *model.GeneralOpenAIRequest) *model.ErrorWithStatusCode {
	// Use existing text helper but we'll intercept and convert the response
	// Create a custom response writer to capture the response
	writer := &responsesResponseWriter{
		ResponseWriter: c.Writer,
		context:        c,
		isStream:       false,
	}
	c.Writer = writer

	// Call existing chat completions pipeline
	err := RelayTextHelper(c)
	if err != nil {
		return err
	}

	return nil
}

// relayResponsesStream handles streaming responses
func relayResponsesStream(c *gin.Context, chatReq *model.GeneralOpenAIRequest) *model.ErrorWithStatusCode {
	// Use existing text helper with stream mode
	writer := &responsesResponseWriter{
		ResponseWriter: c.Writer,
		context:        c,
		isStream:       true,
	}
	c.Writer = writer

	// Call existing chat completions pipeline
	err := RelayTextHelper(c)
	if err != nil {
		return err
	}

	return nil
}

// responsesResponseWriter intercepts the response and converts ChatCompletion to Responses format
type responsesResponseWriter struct {
	gin.ResponseWriter
	context  *gin.Context
	isStream bool
	captured bool
}

func (w *responsesResponseWriter) Write(data []byte) (int, error) {
	// Only convert if we haven't already and if this is responses mode
	if w.captured {
		return w.ResponseWriter.Write(data)
	}

	if !w.isStream {
		// Non-streaming: convert entire response
		var chatResp openai.TextResponse
		err := json.Unmarshal(data, &chatResp)
		if err != nil {
			// If it's an error response, pass through
			logger.Debugf(w.context.Request.Context(), "failed to parse chat response, passing through: %s", err.Error())
			return w.ResponseWriter.Write(data)
		}

		// Convert to Responses format
		responsesResp := convertChatCompletionToResponses(&chatResp)
		respData, err := json.Marshal(responsesResp)
		if err != nil {
			logger.Errorf(w.context.Request.Context(), "failed to marshal responses response: %s", err.Error())
			return w.ResponseWriter.Write(data)
		}

		w.captured = true
		return w.ResponseWriter.Write(respData)
	}

	// Streaming: need to convert each SSE chunk
	return w.writeStreamChunk(data)
}

func (w *responsesResponseWriter) writeStreamChunk(data []byte) (int, error) {
	// Parse SSE format
	lines := strings.Split(string(data), "\n")
	var convertedLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimPrefix(line, "data: ")
			dataStr = strings.TrimSpace(dataStr)

			if dataStr == "[DONE]" {
				convertedLines = append(convertedLines, "data: [DONE]")
				continue
			}

			var chatStreamResp openai.ChatCompletionsStreamResponse
			err := json.Unmarshal([]byte(dataStr), &chatStreamResp)
			if err != nil {
				// Pass through if not parseable
				convertedLines = append(convertedLines, line)
				continue
			}

			// Convert to Responses stream format
			responsesStreamResp := convertChatStreamToResponsesStream(&chatStreamResp)
			respData, err := json.Marshal(responsesStreamResp)
			if err != nil {
				convertedLines = append(convertedLines, line)
				continue
			}

			convertedLines = append(convertedLines, "data: "+string(respData))
		} else {
			convertedLines = append(convertedLines, line)
		}
	}

	convertedData := []byte(strings.Join(convertedLines, "\n"))
	return w.ResponseWriter.Write(convertedData)
}

// convertChatCompletionToResponses converts ChatCompletion response to Responses format
func convertChatCompletionToResponses(chatResp *openai.TextResponse) *model.ResponsesResponse {
	output := make([]model.ResponsesOutputItem, 0, len(chatResp.Choices))

	for _, choice := range chatResp.Choices {
		content := choice.Message.StringContent()
		outputItem := model.ResponsesOutputItem{
			ID:   fmt.Sprintf("msg_%s", uuid.New().String()[:model.ResponsesIDPrefixLength]),
			Type: model.ResponsesOutputTypeMessage,
			Role: choice.Message.Role,
			Content: []model.ResponsesOutputContent{
				{
					Type: model.ResponsesContentTypeOutputText,
					Text: content,
				},
			},
		}
		output = append(output, outputItem)
	}

	return &model.ResponsesResponse{
		ID:      chatResp.Id,
		Object:  "response",
		Created: chatResp.Created,
		Model:   chatResp.Model,
		Output:  output,
		Usage:   &chatResp.Usage,
	}
}

// convertChatStreamToResponsesStream converts streaming ChatCompletion to Responses stream format
func convertChatStreamToResponsesStream(chatResp *openai.ChatCompletionsStreamResponse) *model.ResponsesStreamResponse {
	output := make([]model.ResponsesStreamResponseOutputItem, 0, len(chatResp.Choices))

	for _, choice := range chatResp.Choices {
		var content []model.ResponsesStreamResponseOutputContent
		
		// Get delta content
		deltaContent := choice.Delta.StringContent()
		if deltaContent != "" {
			content = append(content, model.ResponsesStreamResponseOutputContent{
				Type:  model.ResponsesContentTypeOutputText,
				Delta: &deltaContent,
			})
		}

		// Get role if present
		var role *string
		if choice.Delta.Role != "" {
			role = &choice.Delta.Role
		}

		outputItem := model.ResponsesStreamResponseOutputItem{
			Index:   choice.Index,
			Type:    model.ResponsesOutputTypeMessage,
			Role:    role,
			Content: content,
		}

		output = append(output, outputItem)
	}

	return &model.ResponsesStreamResponse{
		ID:      chatResp.Id,
		Object:  "response.delta",
		Created: chatResp.Created,
		Model:   chatResp.Model,
		Output:  output,
		Usage:   chatResp.Usage,
	}
}

// WriteHeader captures headers but delays actual write
func (w *responsesResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responsesResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

var _ gin.ResponseWriter = (*responsesResponseWriter)(nil)
var _ http.Flusher = (*responsesResponseWriter)(nil)
