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
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const API_PREFIX = "/api/v1/"

var apiRequests = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "api_endpoints_requests",
	Help: "The total number requests per API endpoint",
}, []string{"url"})

var apiResponses = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "response_time",
	Help:    "Response time",
	Buckets: prometheus.LinearBuckets(0, 20, 20),
}, []string{"url"})

type Status struct {
	Status string `json:"status"`
}

var OkStatus = Status{
	Status: "ok",
}

func countEndpoint(request *http.Request, start time.Time) {
	url := request.URL.String()
	duration := time.Since(start)
	log.Printf("Time to serve the page: %s\n", duration)

	apiRequests.With(prometheus.Labels{"url": url}).Inc()

	apiResponses.With(prometheus.Labels{"url": url}).Observe(float64(duration.Microseconds()))
}

func retrieveIdRequestParameter(request *http.Request) (int64, error) {
	id_var := mux.Vars(request)["id"]
	return strconv.ParseInt(id_var, 10, 0)
}

func mainEndpoint(writer http.ResponseWriter, request *http.Request) {
	start := time.Now()
	io.WriteString(writer, "Hello world!\n")
	countEndpoint(request, start)
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

func Initialize(address, ldap string) {
	log.Println("Initializing HTTP server at", ldap)
	router := mux.NewRouter().StrictSlash(true)
	router.Use(logRequest)
	router.Use(JwtAuthentication)

	router.HandleFunc(API_PREFIX, mainEndpoint).Methods("GET")
	router.HandleFunc(API_PREFIX+"login", func(w http.ResponseWriter, r *http.Request) { login(w, r, ldap) }).Methods("POST")

	clientRouter := router.PathPrefix(API_PREFIX + "client").Subrouter()
	clientRouter.HandleFunc("/cluster", func(w http.ResponseWriter, r *http.Request) { getClusters(w, r) }).Methods("GET")
	log.Println("Starting HTTP server at", ldap)

	var err error

	err = http.ListenAndServe(address, router)
	if err != nil {
		log.Fatal("Unable to initialize HTTP server", err)
		os.Exit(2)
	}
}
