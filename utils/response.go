package utils

import (
	"encoding/json"
	"net/http"
)

// BuildResponse - build response for RestAPI request
func BuildResponse(status string) map[string]interface{} {
	return map[string]interface{}{"status": status}
}

func addJSONHeader(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
}

// SendResponse - return JSON response
func SendResponse(w http.ResponseWriter, data map[string]interface{}) {
	addJSONHeader(w)
	json.NewEncoder(w).Encode(data)
}

// SendError - return error response
func SendError(w http.ResponseWriter, data map[string]interface{}) {
	w.WriteHeader(http.StatusBadRequest)
	SendResponse(w, data)
}

// SendForbidden - return response with status Forbidden
func SendForbidden(w http.ResponseWriter, data map[string]interface{}) {
	w.WriteHeader(http.StatusForbidden)
	SendResponse(w, data)
}
