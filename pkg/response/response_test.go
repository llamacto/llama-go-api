package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSuccess(t *testing.T) {
	// 创建测试路由
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 测试数据
	testData := map[string]interface{}{
		"id":   1,
		"name": "Test User",
	}

	// 调用 Success
	Success(c, testData)

	// 验证状态码
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// 验证响应体
	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		t.Errorf("Expected code 0, got %d", response.Code)
	}

	if response.Message != "success" {
		t.Errorf("Expected message 'success', got '%s'", response.Message)
	}

	if response.Data == nil {
		t.Fatal("Expected data to be present")
	}
}

func TestSuccess_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, nil)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		t.Errorf("Expected code 0, got %d", response.Code)
	}

	if response.Message != "success" {
		t.Errorf("Expected message 'success', got '%s'", response.Message)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testCode := http.StatusBadRequest
	testMessage := "Invalid input"

	Error(c, testCode, testMessage)

	// 验证状态码
	if w.Code != testCode {
		t.Errorf("Expected status code %d, got %d", testCode, w.Code)
	}

	// 验证响应体
	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != testCode {
		t.Errorf("Expected code %d, got %d", testCode, response.Code)
	}

	if response.Message != testMessage {
		t.Errorf("Expected message '%s', got '%s'", testMessage, response.Message)
	}

	if response.Data != nil {
		t.Error("Expected data to be nil for error response")
	}
}

func TestError_DifferentStatusCodes(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		message    string
	}{
		{"BadRequest", http.StatusBadRequest, "Bad request"},
		{"Unauthorized", http.StatusUnauthorized, "Unauthorized"},
		{"Forbidden", http.StatusForbidden, "Forbidden"},
		{"NotFound", http.StatusNotFound, "Not found"},
		{"InternalError", http.StatusInternalServerError, "Internal server error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Error(c, tc.statusCode, tc.message)

			if w.Code != tc.statusCode {
				t.Errorf("Expected status code %d, got %d", tc.statusCode, w.Code)
			}

			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Code != tc.statusCode {
				t.Errorf("Expected code %d, got %d", tc.statusCode, response.Code)
			}

			if response.Message != tc.message {
				t.Errorf("Expected message '%s', got '%s'", tc.message, response.Message)
			}
		})
	}
}

func TestResponse_JSONStructure(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := map[string]string{
		"key": "value",
	}

	Success(c, testData)

	// 验证 JSON 结构
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &jsonMap); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// 检查必需字段
	if _, ok := jsonMap["code"]; !ok {
		t.Error("Response missing 'code' field")
	}

	if _, ok := jsonMap["message"]; !ok {
		t.Error("Response missing 'message' field")
	}

	if _, ok := jsonMap["data"]; !ok {
		t.Error("Response missing 'data' field")
	}
}

// Benchmark tests
func BenchmarkSuccess(b *testing.B) {
	gin.SetMode(gin.TestMode)
	data := map[string]string{"test": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		Success(c, data)
	}
}

func BenchmarkError(b *testing.B) {
	gin.SetMode(gin.TestMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		Error(c, http.StatusBadRequest, "test error")
	}
}
