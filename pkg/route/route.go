package route

import (
	"io"
	"net/http"
)

// Healthz simple healthcheck endpoint
// HTTP/200
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}
