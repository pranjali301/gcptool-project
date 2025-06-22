// function/function.go
package function

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("HandleRequest", HandleRequest)
}

// HandleRequest processes HTTP requests - under 20 lines as required
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}

	response := map[string]interface{}{
		"message": fmt.Sprintf("Hello, %s! From GCP Cloud Function", name),
		"time":    time.Now().Format(time.RFC3339),
		"method":  r.Method,
	}
	json.NewEncoder(w).Encode(response)
}
