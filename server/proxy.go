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

package server

// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-operator-ldapauth/packages/server/proxy.html

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// HandleHTTP handle all routes, used for proxying them to controller
func (s *Server) HandleHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("Proxying to", s.Proxy)
	tempURL, _ := url.Parse(s.Proxy)
	req.URL.Host = tempURL.Host
	req.URL.Scheme = tempURL.Scheme
	req.URL.Path = strings.Replace(req.URL.Path, APIPrefix, s.ProxyPrefix, 1)
	req.Host = tempURL.Host
	resp, err := s.Transport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println("Unable to close response body")
		}
	}()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Println("Error copying response body")
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
