# /v1/responses Endpoint Implementation

## Overview
This implementation adds support for the OpenAI Responses API endpoint (`/v1/responses`) to One-API. The implementation follows a minimal, pragmatic approach by reusing the existing chat completions pipeline internally.

## Architecture

### Request Flow
1. Client sends request to `/v1/responses`
2. Request is authenticated and distributed (middleware)
3. `RelayResponsesHelper` converts the request:
   - Parses `ResponsesRequest` 
   - Converts `input` field to `messages` format
   - Creates internal `GeneralOpenAIRequest` (chat completion format)
4. Request flows through existing text relay pipeline:
   - Token counting
   - Quota management
   - Channel selection
   - Provider adaptation
5. Response intercepted and converted back to Responses format
6. Client receives `ResponsesResponse`

### Key Components

#### 1. Relay Mode (`relay/relaymode/`)
- Added `Responses` constant to `define.go`
- Updated `GetByPath()` to recognize `/v1/responses` path

#### 2. Models (`relay/model/responses.go`)
New structs:
- `ResponsesRequest`: Input request format
  - `model` (required): Model name
  - `input` (string or array): User input
  - `messages` (optional): Pre-formatted messages
  - `stream` (bool): Enable streaming
  - Additional fields: `max_tokens`, `temperature`, `top_p`

- `ResponsesResponse`: Non-streaming response
  ```json
  {
    "id": "chatcmpl-xxx",
    "object": "response",
    "created": 1234567890,
    "model": "gpt-3.5-turbo",
    "output": [
      {
        "id": "msg_xxx",
        "type": "message",
        "role": "assistant",
        "content": [
          {
            "type": "output_text",
            "text": "Response text here"
          }
        ]
      }
    ],
    "usage": {...}
  }
  ```

- `ResponsesStreamResponse`: Streaming response with delta format

#### 3. Controller (`relay/controller/responses.go`)
- `RelayResponsesHelper()`: Main handler
  - Parses and validates request
  - Converts input to messages
  - Delegates to existing text relay pipeline
  
- `responsesResponseWriter`: Custom response writer
  - Intercepts ChatCompletion responses
  - Converts to Responses format on-the-fly
  - Handles both streaming and non-streaming

#### 4. Validation & Helpers
- Updated `ValidateTextRequest()` to handle Responses mode
- Updated `getPromptTokens()` to count tokens for Responses mode

#### 5. Router (`router/relay.go`)
- Added route: `POST /v1/responses`

## API Usage

### Non-Streaming Example

**Request:**
```bash
curl -X POST http://localhost:3000/v1/responses \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "input": "Hello, how are you?"
  }'
```

**Response:**
```json
{
  "id": "chatcmpl-xxx",
  "object": "response",
  "created": 1703472000,
  "model": "gpt-3.5-turbo",
  "output": [
    {
      "id": "msg_abc123",
      "type": "message",
      "role": "assistant",
      "content": [
        {
          "type": "output_text",
          "text": "Hello! I'm doing well, thank you for asking. How can I assist you today?"
        }
      ]
    }
  ],
  "usage": {
    "prompt_tokens": 13,
    "completion_tokens": 20,
    "total_tokens": 33
  }
}
```

### Streaming Example

**Request:**
```bash
curl -X POST http://localhost:3000/v1/responses \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "input": "Tell me a joke",
    "stream": true
  }'
```

**Response (SSE format):**
```
data: {"id":"chatcmpl-xxx","object":"response.delta","created":1703472000,"model":"gpt-3.5-turbo","output":[{"index":0,"type":"message","role":"assistant","content":[{"type":"output_text","delta":"Why"}]}]}

data: {"id":"chatcmpl-xxx","object":"response.delta","created":1703472000,"model":"gpt-3.5-turbo","output":[{"index":0,"type":"message","content":[{"type":"output_text","delta":" did"}]}]}

data: [DONE]
```

### Using Messages Format

**Request:**
```bash
curl -X POST http://localhost:3000/v1/responses \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "What is the capital of France?"}
    ]
  }'
```

## Implementation Details

### Input Conversion
The `ParseInput()` method handles different input formats:
- **String**: Converted to single user message
- **Array of strings**: Each string becomes a user message
- **Messages array**: Used directly without conversion

### Response Conversion
Non-streaming responses are converted by:
1. Extracting assistant message content from ChatCompletion
2. Wrapping in `output` array with proper structure
3. Changing `object` field to "response"

Streaming responses are converted by:
1. Parsing SSE chunks
2. Converting delta content to Responses stream format
3. Maintaining proper field structure for each chunk

### Compatibility
- ✅ Reuses all existing middleware (auth, rate limiting, etc.)
- ✅ Works with all OpenAI-compatible providers
- ✅ Quota management fully functional
- ✅ Channel selection and retry logic preserved
- ✅ Logging and monitoring work as expected

### Limitations (MVP)
- Tool calling not yet supported
- Multimodal content not yet supported
- Advanced Responses API features not implemented
- Focus on basic text input/output

## Testing

Unit tests in `relay/controller/responses_test.go`:
- Request parsing for different input formats
- Validation of required fields
- Stream flag handling

Run tests:
```bash
go test ./relay/controller/... -v
```

## Edge Cases Handled

1. **Empty input**: Returns validation error
2. **Missing model**: Returns validation error
3. **Both input and messages provided**: Messages take precedence
4. **Invalid JSON**: Standard error handling
5. **Provider errors**: Passed through with proper error format

## Future Enhancements

Potential improvements for future iterations:
1. Support for tool calling
2. Multimodal content support
3. More advanced Responses API features
4. Performance optimizations for streaming
5. Additional validation rules

## Notes

- This implementation prioritizes **pragmatic compatibility** over 100% spec compliance
- The goal is to provide a working endpoint that integrates seamlessly with One-API's architecture
- All existing functionality remains unchanged and unaffected
