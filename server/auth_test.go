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

package server

// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-operator-ldapauth/packages/server/auth_test.html

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testServer = &Server{
	Address:   ":8081",
	LDAP:      "",
	Proxy:     "http://localhost:8080",
	Transport: http.DefaultTransport,
}

func TestServerLoginUnauthorized(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{"login": "tester", "password": "1234"})
	req, err := http.NewRequest("POST", APIPrefix+"login", bytes.NewBuffer(body))
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testServer.Login)
	handler.ServeHTTP(rr, req)

	// Check content-type header
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type header value: %v but got %v",
			"application/json; charset=utf-8", contentType)
	}

	// Check the status code
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Expected status code: %v but got %v",
			http.StatusUnauthorized, status)
	}

}

func TestJwtAuthNoProxy(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{"login": "tester", "password": "1234"})
	req, err := http.NewRequest("POST", APIPrefix+"login", bytes.NewBuffer(body))
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testServer.HandleHTTP)
	newHandler := testServer.JWTAuthentication(handler)
	newHandler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusServiceUnavailable {
		t.Errorf("Expected status code: %v but got %v",
			http.StatusServiceUnavailable, status)
	}

}
