package function

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleRequest(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantName   string
		wantMethod string
	}{
		{
			name:       "No name parameter",
			url:        "/",
			wantName:   "World",
			wantMethod: "GET",
		},
		{
			name:       "With name parameter",
			url:        "/?name=Alice",
			wantName:   "Alice",
			wantMethod: "GET",
		},
		{
			name:       "Empty name parameter",
			url:        "/?name=",
			wantName:   "World",
			wantMethod: "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			rr := httptest.NewRecorder()

			HandleRequest(rr, req)

			// Check status code
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			// Check content type
			expected := "application/json"
			if ct := rr.Header().Get("Content-Type"); ct != expected {
				t.Errorf("handler returned wrong content type: got %v want %v",
					ct, expected)
			}

			// Check response body
			var response map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Errorf("Failed to decode response: %v", err)
			}

			// Check message contains expected name
			if message, ok := response["message"].(string); ok {
				expectedMsg := "Hello, " + tt.wantName + "!"
				if !contains(message, expectedMsg) {
					t.Errorf("Expected message to contain '%s', got '%s'", expectedMsg, message)
				}
			} else {
				t.Error("Expected 'message' field in response")
			}

			// Check method
			if method, ok := response["method"].(string); ok {
				if method != tt.wantMethod {
					t.Errorf("Expected method '%s', got '%s'", tt.wantMethod, method)
				}
			} else {
				t.Error("Expected 'method' field in response")
			}

			// Check time field exists
			if _, ok := response["time"]; !ok {
				t.Error("Expected 'time' field in response")
			}
		})
	}
}

func TestHandleRequestPOST(t *testing.T) {
	req := httptest.NewRequest("POST", "/?name=Bob", nil)
	rr := httptest.NewRecorder()

	HandleRequest(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if method, ok := response["method"].(string); ok {
		if method != "POST" {
			t.Errorf("Expected method 'POST', got '%s'", method)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
