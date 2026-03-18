package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type errorResponse struct {
	Error string `json:"error"`
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, errorResponse{Error: msg})
}

func extractID(r *http.Request, w http.ResponseWriter) (uint, bool) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var rawID string
	for i, p := range parts {
		if p == "departments" && i+1 < len(parts) {
			rawID = parts[i+1]
			break
		}
	}
	v, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid department id")
		return 0, false
	}
	return uint(v), true
}
