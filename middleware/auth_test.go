package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestShouldCheckModel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "should check model for /v1/completions",
			path:     "/v1/completions",
			expected: true,
		},
		{
			name:     "should check model for /v1/chat/completions",
			path:     "/v1/chat/completions",
			expected: true,
		},
		{
			name:     "should check model for /v1/responses",
			path:     "/v1/responses",
			expected: true,
		},
		{
			name:     "should check model for /v1/images/generations",
			path:     "/v1/images/generations",
			expected: true,
		},
		{
			name:     "should check model for /v1/audio/transcriptions",
			path:     "/v1/audio/transcriptions",
			expected: true,
		},
		{
			name:     "should check model for /v1/audio/translations",
			path:     "/v1/audio/translations",
			expected: true,
		},
		{
			name:     "should not check model for /v1/models",
			path:     "/v1/models",
			expected: false,
		},
		{
			name:     "should not check model for /v1/embeddings",
			path:     "/v1/embeddings",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", tt.path, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			result := shouldCheckModel(c)
			if result != tt.expected {
				t.Errorf("shouldCheckModel() for path %s = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}
