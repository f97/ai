package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/model"
)

func TestResponsesModelValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize test database
	model.InitDB()

	tests := []struct {
		name           string
		requestBody    map[string]any
		expectedStatus int
		checkError     bool
		errorContains  string
	}{
		{
			name: "valid request with model",
			requestBody: map[string]any{
				"model": "gpt-3.5-turbo",
				"input": "Hello",
			},
			expectedStatus: http.StatusOK, // Will fail at distribute stage but that's ok - we just test auth passes
			checkError:     false,
		},
		{
			name: "missing model field - should return 400",
			requestBody: map[string]any{
				"input": "Hello",
			},
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
			errorContains:  "model",
		},
		{
			name: "empty model field - should return 400",
			requestBody: map[string]any{
				"model": "",
				"input": "Hello",
			},
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
			errorContains:  "model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test router
			router := gin.New()
			
			// Add the auth middleware
			router.Use(func(c *gin.Context) {
				// Mock a valid token check - set user ID so auth passes
				c.Set(ctxkey.Id, 1)
				c.Next()
			})

			// Add the model checking logic
			router.POST("/v1/responses", func(c *gin.Context) {
				// Simulate what TokenAuth does
				requestModel, err := getRequestModel(c)
				if err != nil && shouldCheckModel(c) {
					abortWithMessage(c, http.StatusBadRequest, err.Error())
					return
				}
				c.Set(ctxkey.RequestModel, requestModel)
				
				// If we reach here, validation passed
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			// Create request
			jsonData, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Check error message if expected
			if tt.checkError {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse error response: %v", err)
					return
				}

				if errorObj, ok := response["error"].(map[string]any); ok {
					if message, ok := errorObj["message"].(string); ok {
						t.Logf("Error message: %s", message)
					} else {
						t.Error("Error message not found in response")
					}
				} else {
					t.Error("Error object not found in response")
				}
			}
		})
	}
}
