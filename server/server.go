/*
Copyright © 2019, 2020, 2021, 2022 Red Hat, Inc.

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

// Package server contains implementation of REST API server (HTTPServer) for the
// LDAP proxy.
//
// Generated documentation is available at:
// https://godoc.org/github.com/RedHatInsights/insights-operator-ldapauth/server
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-operator-ldapauth/packages/server/server.html
package server

import (
	"github.com/RedHatInsights/insights-operator-utils/env"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

// APIPrefix is appended before all REST API endpoint addresses
var APIPrefix = env.GetEnv("CONTROLLER_PREFIX", "/api/v1/")

// Server basic configuration of server
type Server struct {
	Address     string
	LDAP        string
	Proxy       string
	ProxyPrefix string
	Transport   http.RoundTripper
}

// Status structure for RestAPI response
type Status struct {
	Status string `json:"status"`
}

// OkStatus prepared successful response
var OkStatus = Status{
	Status: "ok",
}

func logRequestHandler(writer http.ResponseWriter, request *http.Request, nextHandler http.Handler) {
	log.Println("Request URI: " + request.RequestURI)
	log.Println("Request method: " + request.Method)
	nextHandler.ServeHTTP(writer, request)
}

func logRequest(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			logRequestHandler(writer, request, nextHandler)
		})
}

func (s *Server) addDefaultHeaders(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			nextHandler.ServeHTTP(writer, request)
		})
}

// Initialize main function that start server
func (s *Server) Initialize() {
	log.Println("API Prefix: ", APIPrefix)
	router := mux.NewRouter().StrictSlash(true)
	router.Use(logRequest)
	router.Use(s.JWTAuthentication)
	router.Use(s.addDefaultHeaders)

	router.HandleFunc(APIPrefix+"login", s.Login).Methods("POST")
	router.PathPrefix(APIPrefix).Handler(http.HandlerFunc(s.HandleHTTP))

	log.Println("Starting HTTP server at", s.Address)

	err := http.ListenAndServe(s.Address, router) // #nosec G114
	if err != nil {
		log.Fatal("Unable to initialize HTTP server", err)
		os.Exit(2)
	}
}
