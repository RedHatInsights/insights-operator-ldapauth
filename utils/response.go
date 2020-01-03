// Utils for REST API

/*
Copyright © 2019, 2020 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"encoding/json"
	"net/http"
)

const (
	contentType = "Content-Type"
	appJSON     = "application/json; charset=utf-8"
)

// BuildResponse builds response for RestAPI request
func BuildResponse(status string) map[string]interface{} {
	return map[string]interface{}{"status": status}
}

// SendResponse returns JSON response
func SendResponse(w http.ResponseWriter, data map[string]interface{}) {
	json.NewEncoder(w).Encode(data)
}

// SendError returns error response
func SendError(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Set(contentType, appJSON)
	w.WriteHeader(http.StatusBadRequest)
	SendResponse(w, data)
}

// SendForbidden returns response with HTTP status Forbidden
func SendForbidden(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Set(contentType, appJSON)
	w.WriteHeader(http.StatusForbidden)
	SendResponse(w, data)
}

// SendUnauthorized returns error response for unauthorized access
func SendUnauthorized(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Set(contentType, appJSON)
	w.WriteHeader(http.StatusUnauthorized)
	SendResponse(w, data)
}
