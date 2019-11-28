package utils

import (
	"encoding/json"
	"net/http"
)

func BuildResponse(status string) map[string]interface{} {
	return map[string]interface{}{"status": status}
}

func addJsonHeader(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
}

func SendResponse(w http.ResponseWriter, data map[string]interface{}) {
	addJsonHeader(w)
	json.NewEncoder(w).Encode(data)
}

func SendError(w http.ResponseWriter, data map[string]interface{}) {
	w.WriteHeader(http.StatusBadRequest)
	SendResponse(w, data)
}

func SendForbidden(w http.ResponseWriter, data map[string]interface{}) {
	w.WriteHeader(http.StatusForbidden)
	SendResponse(w, data)
}
