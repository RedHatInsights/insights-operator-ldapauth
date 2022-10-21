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

// Entry point to the LDAP proxy service.
//
// That service contains implementation of authentication mechanism used to
// validate user credentials and create auth token.
//
// Generated documentation is available at:
// https://godoc.org/github.com/RedHatInsights/insights-operator-ldapauth/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-operator-ldapauth/packages/main.html
package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/RedHatInsights/insights-operator-utils/env"
	"github.com/redhatinsights/insights-operator-ldapauth/server"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func createTLS(tlsCert, tlsKey string) http.RoundTripper {
	// disable "G304 (CWE-22): Potential file inclusion via variable"
	caCert, err := os.ReadFile(tlsCert)
	if err != nil {
		log.Fatal(err)
	}
	cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	// Create a HTTPS client and supply the created CA pool
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      caCertPool,
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}
	return transport
}

func main() {
	// parse the configuration
	configFile := env.GetEnv("INSIGHTS_CONTROLLER_CONFIG_FILE", "./config")
	// we need to separate the directory name and filename without extension
	directory, basename := filepath.Split(configFile)
	file := strings.TrimSuffix(basename, filepath.Ext(basename))
	// parse the configuration
	viper.SetConfigName(file)
	viper.AddConfigPath(directory)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	flag.Parse()

	serviceCfg := viper.Sub("service")
	s := server.Server{
		LDAP:        serviceCfg.GetString("ldap"),
		Address:     serviceCfg.GetString("address"),
		Proxy:       serviceCfg.GetString("proxy"),
		ProxyPrefix: serviceCfg.GetString("proxy_prefix"),
	}

	if serviceCfg.GetBool("proxy_tls") {
		s.Transport = createTLS(serviceCfg.GetString("tls_cert"), serviceCfg.GetString("tls_key"))
	} else {
		s.Transport = http.DefaultTransport
	}

	s.Initialize()
}
