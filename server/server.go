/*
Copyright © 2019 Red Hat, Inc.

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

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
	"net/http"
	"os"
)

// APIPrefix is appended before all REST API endpoint addresses
const APIPrefix = "/api/v1/"

// Server basic configuration of server
type Server struct {
	Address string
	LDAP    string
	Proxy   string
}

var apiRequests = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "api_endpoints_requests",
	Help: "The total number requests per API endpoint",
}, []string{"url"})

var apiResponses = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "response_time",
	Help:    "Response time",
	Buckets: prometheus.LinearBuckets(0, 20, 20),
}, []string{"url"})

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

func (s Server) addDefaultHeaders(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			nextHandler.ServeHTTP(writer, request)
		})
}

// Initialize main function that start server
func (s Server) Initialize() {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(logRequest)
	router.Use(s.JWTAuthentication)
	router.Use(s.addDefaultHeaders)

	router.HandleFunc(APIPrefix+"login", s.Login).Methods("POST")
	router.PathPrefix(APIPrefix).Handler(http.HandlerFunc(s.HandleHTTP))

	log.Println("Starting HTTP server at", s.Address)

	var err error

	err = http.ListenAndServe(s.Address, router)
	if err != nil {
		log.Fatal("Unable to initialize HTTP server", err)
		os.Exit(2)
	}
}
