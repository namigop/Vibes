package httpapi

import (
	"net/http"
)

// HealthHandler returns 200 OK with a small JSON payload.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
