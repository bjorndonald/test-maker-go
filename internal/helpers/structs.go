package helpers

// RequestBody represents the payload sent to OpenAI API
type EmbeddingRequestBody struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// ResponseBody represents the response from OpenAI API
type EmbeddingResponseBody struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}
