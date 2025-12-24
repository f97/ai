# Architecture Diagram: /v1/responses Endpoint

## High-Level Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                           Client Application                         │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             │ POST /v1/responses
                             │ { "model": "gpt-3.5-turbo", "input": "..." }
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Router (router/relay.go)                        │
│  - CORS Middleware                                                   │
│  - Gzip Decode Middleware                                            │
│  - Token Auth Middleware ✓                                          │
│  - Distribute Middleware (channel selection) ✓                      │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│             Relay Controller (controller/relay.go)                   │
│  - Get relay mode: relaymode.Responses                               │
│  - Route to RelayResponsesHelper                                     │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│        Responses Controller (relay/controller/responses.go)          │
│                                                                       │
│  1. Parse ResponsesRequest                                           │
│     - Validate model (required)                                      │
│     - Parse input or messages                                        │
│                                                                       │
│  2. Convert to Chat Format                                           │
│     ┌─────────────────────────────┐                                 │
│     │ if input is string:         │                                 │
│     │   → [{"role": "user",       │                                 │
│     │      "content": input}]     │                                 │
│     │                             │                                 │
│     │ if input is array:          │                                 │
│     │   → multiple user messages  │                                 │
│     │                             │                                 │
│     │ if messages provided:       │                                 │
│     │   → use directly            │                                 │
│     └─────────────────────────────┘                                 │
│                                                                       │
│  3. Create GeneralOpenAIRequest                                      │
│     - model, messages, stream, max_tokens, etc.                      │
│                                                                       │
│  4. Wrap with responsesResponseWriter                                │
│     - Intercepts response                                            │
│     - Converts format on-the-fly                                     │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│        Text Relay Helper (relay/controller/text.go)                  │
│                                                                       │
│  ✓ Validate request                                                  │
│  ✓ Model name mapping                                                │
│  ✓ System prompt handling                                            │
│  ✓ Calculate model ratio                                             │
│  ✓ Count prompt tokens ←─── getPromptTokens() handles Responses     │
│  ✓ Pre-consume quota                                                 │
│  ✓ Get adaptor (OpenAI, Claude, etc.)                                │
│  ✓ Convert request format (if needed)                                │
│  ✓ Make provider request                                             │
│  ✓ Handle provider response                                          │
│  ✓ Post-consume quota                                                │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Provider (OpenAI, etc.)                           │
│  - Returns ChatCompletion format response                            │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│          responsesResponseWriter (intercepts response)               │
│                                                                       │
│  Non-Streaming:                    Streaming:                        │
│  ┌──────────────────────┐         ┌──────────────────────┐          │
│  │ Full response buffer │         │ SSE chunk by chunk   │          │
│  │ Parse TextResponse   │         │ Parse each delta     │          │
│  │ Convert to Responses │         │ Convert to Responses │          │
│  │ Write to client      │         │ Stream to client     │          │
│  └──────────────────────┘         └──────────────────────┘          │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                           Client Application                         │
│  Receives ResponsesResponse format:                                  │
│  {                                                                    │
│    "id": "chatcmpl-xxx",                                             │
│    "object": "response",                                             │
│    "output": [{                                                      │
│      "id": "msg_xxx",                                                │
│      "type": "message",                                              │
│      "role": "assistant",                                            │
│      "content": [{"type": "output_text", "text": "..."}]            │
│    }],                                                               │
│    "usage": {...}                                                    │
│  }                                                                    │
└─────────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Request Conversion
```
ResponsesRequest               GeneralOpenAIRequest
┌──────────────┐              ┌──────────────┐
│ model        │──────────────>│ model        │
│ input        │─┐            │              │
│ messages     │ │ ParseInput()│ messages     │
│ stream       │─┴───────────>│ stream       │
│ max_tokens   │──────────────>│ max_tokens   │
│ temperature  │──────────────>│ temperature  │
└──────────────┘              └──────────────┘
```

### 2. Response Conversion (Non-Streaming)
```
ChatCompletion Response        Responses Response
┌──────────────────┐          ┌──────────────────┐
│ id               │─────────>│ id               │
│ object: "chat... │          │ object: "response│
│ choices: [       │          │ output: [        │
│   {              │          │   {              │
│     message: {   │─────────>│     type: "msg"  │
│       role       │          │     role         │
│       content    │─────────>│     content: [{  │
│     }            │          │       type: "out │
│     finish_...   │          │       text: ...  │
│   }              │          │     }]           │
│ ]                │          │   }              │
│ usage            │─────────>│ ]                │
│                  │          │ usage            │
└──────────────────┘          └──────────────────┘
```

### 3. Response Conversion (Streaming)
```
SSE Chunks (ChatCompletion)    SSE Chunks (Responses)
┌───────────────────────┐     ┌───────────────────────┐
│ data: {               │     │ data: {               │
│   choices: [{         │     │   output: [{          │
│     delta: {          │────>│     type: "message"   │
│       content: "Hi"   │     │     content: [{       │
│     }                 │     │       type: "output_  │
│   }]                  │     │       delta: "Hi"     │
│ }                     │     │     }]                │
│                       │     │   }]                  │
└───────────────────────┘     └───────────────────────┘
```

## Key Design Decisions

### ✅ Advantages of This Architecture

1. **Minimal Code Changes**
   - Only 1,098 lines added across 13 files
   - No changes to core pipeline logic
   - No duplicate auth/quota/logging code

2. **Maximum Reuse**
   - 100% reuse of existing middleware
   - 100% reuse of text relay pipeline
   - 100% reuse of provider adaptors
   - 100% reuse of quota management

3. **Clean Separation**
   - Request conversion: isolated in responses.go
   - Response conversion: isolated in responsesResponseWriter
   - No pollution of shared code

4. **Easy Testing**
   - Unit tests for conversion logic
   - Integration tests reuse existing test infrastructure
   - No mocking required for most tests

5. **Maintainability**
   - Changes to chat completions automatically benefit responses
   - Bug fixes in pipeline benefit both endpoints
   - Single source of truth for provider logic

### ⚠️ Trade-offs

1. **Response Interception**
   - Uses custom ResponseWriter to intercept
   - Adds small overhead for parsing/conversion
   - Acceptable for MVP scope

2. **Format Limitations**
   - Simplified Responses format vs full spec
   - Some advanced features not supported
   - Good enough for pragmatic compatibility

3. **Streaming Complexity**
   - SSE parsing and conversion per chunk
   - More complex than buffered approach
   - But necessary for real streaming

## Files Organization

```
/home/runner/work/1ai/1ai/
├── relay/
│   ├── relaymode/
│   │   ├── define.go         ← Added Responses constant
│   │   └── helper.go         ← Added path matching
│   ├── model/
│   │   ├── constant.go       ← Added Responses constants
│   │   └── responses.go      ← NEW: Response models
│   └── controller/
│       ├── responses.go       ← NEW: Main controller
│       ├── responses_test.go  ← NEW: Unit tests
│       ├── helper.go          ← Updated token counting
│       └── validator/
│           └── validation.go  ← Updated validation
├── controller/
│   └── relay.go              ← Added Responses case
├── router/
│   └── relay.go              ← Added route
├── docs/
│   ├── responses-api-implementation.md  ← NEW: Detailed docs
│   └── responses-api-summary.md         ← NEW: Summary
└── bin/
    └── test_responses_api.sh ← NEW: Test script
```

## Deployment Flow

```
┌────────────────┐
│ Build Binary   │
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ Run Tests      │
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ Security Check │
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ Deploy         │  ← No config changes needed
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ Ready!         │  ← /v1/responses available
└────────────────┘
```
