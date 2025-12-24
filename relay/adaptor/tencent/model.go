package tencent

type Message struct {
	Role    string `json:"Role"`
	Content string `json:"Content"`
}

type ChatRequest struct {
	// Model name, options include hunyuan-lite, hunyuan-standard, hunyuan-standard-256K, hunyuan-pro.
	// For model introduction, please read [Product Overview](https://cloud.tencent.com/document/product/1729/104753).
	//
	// Note:
	// Different models have different billing, please refer to [Purchase Guide](https://cloud.tencent.com/document/product/1729/97731).
	Model *string `json:"Model"`
	// Chat context information.
	// Note:
	// 1. Maximum length is 40, arranged in the array from oldest to newest by conversation time.
	// 2. Message.Role options: system, user, assistant.
	// 其中，system 角色可选，如存在则必须位于列表的最开始。user 和 assistant 需交替出现（一问一答），以 user 提问开始和结束，且 Content 不能为空。Role 的顺序示例：[system（可选） user assistant user assistant user ...]。
	// 3. The total length of Content in Messages cannot exceed the model input length limit (refer to [Product Overview](https://cloud.tencent.com/document/product/1729/104753) document). If exceeded, the beginning content will be truncated, keeping only the tail content.
	Messages []*Message `json:"Messages"`
	// Streaming call switch.
	// Note:
	// 1. Default is non-streaming call (false) when no value is passed.
	// 2. When streaming, results are returned incrementally via SSE protocol (return value is taken from Choices[n].Delta, incremental data needs to be concatenated to get complete results).
	// 3. For non-streaming calls:
	// Calling method is the same as regular HTTP requests.
	// Interface response takes a long time, **set to true for lower latency**.
	// Only returns the final result once (return value is taken from Choices[n].Message).
	//
	// Note:
	// When calling through SDK, streaming and non-streaming calls require **different methods** to get return values. Refer to the comments or examples in the SDK (in the examples/hunyuan/v20230901/ directory of each language SDK code repository).
	Stream *bool `json:"Stream"`
	// Note:
	// 1. Affects the diversity of output text. The larger the value, the stronger the diversity of generated text.
	// 2. Value range is [0.0, 1.0]. When no value is passed, the recommended value for each model is used.
	// 3. Not recommended unless necessary. Unreasonable values will affect the results.
	TopP *float64 `json:"TopP,omitempty"`
	// Note:
	// 1. 较高的数值会使输出更加随机，而较低的数值会使其更加集中和确定。
	// 2. Value range is [0.0, 2.0]. When no value is passed, the recommended value for each model is used.
	// 3. Not recommended unless necessary. Unreasonable values will affect the results.
	Temperature *float64 `json:"Temperature,omitempty"`
}

type Error struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type Usage struct {
	PromptTokens     int `json:"PromptTokens"`
	CompletionTokens int `json:"CompletionTokens"`
	TotalTokens      int `json:"TotalTokens"`
}

type ResponseChoices struct {
	FinishReason string  `json:"FinishReason,omitempty"` // 流式结束标志位，为 stop 则表示尾包
	Messages     Message `json:"Message,omitempty"`      // 内容，同步模式返回内容，流模式为 null 输出 content 内容总数最多支持 1024token。
	Delta        Message `json:"Delta,omitempty"`        // 内容，流模式返回内容，同步模式为 null 输出 content 内容总数最多支持 1024token。
}

type ChatResponse struct {
	Choices []ResponseChoices `json:"Choices,omitempty"`   // 结果
	Created int64             `json:"Created,omitempty"`   // unix 时间戳的字符串
	Id      string            `json:"Id,omitempty"`        // 会话 id
	Usage   Usage             `json:"Usage,omitempty"`     // token 数量
	Error   Error             `json:"Error,omitempty"`     // 错误信息 注意：此字段可能返回 null，表示取不到有效值
	Note    string            `json:"Note,omitempty"`      // 注释
	ReqID   string            `json:"RequestId,omitempty"` // 唯一请求 Id，每次请求都会返回。用于反馈接口入参
}

type ChatResponseP struct {
	Response ChatResponse `json:"Response,omitempty"`
}

type EmbeddingRequest struct {
	InputList []string `json:"InputList"`
}

type EmbeddingData struct {
	Embedding []float64 `json:"Embedding"`
	Index     int       `json:"Index"`
	Object    string    `json:"Object"`
}

type EmbeddingUsage struct {
	PromptTokens int `json:"PromptTokens"`
	TotalTokens  int `json:"TotalTokens"`
}

type EmbeddingResponse struct {
	Data           []EmbeddingData `json:"Data"`
	EmbeddingUsage EmbeddingUsage  `json:"Usage,omitempty"`
	RequestId      string          `json:"RequestId,omitempty"`
	Error          Error           `json:"Error,omitempty"`
}

type EmbeddingResponseP struct {
	Response EmbeddingResponse `json:"Response,omitempty"`
}
