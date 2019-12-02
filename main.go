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
package main

import (
	"flag"
	"fmt"
	"github.com/redhatinsights/insights-operator-ldapauth/server"
	"github.com/spf13/viper"
)

func main() {
	// parse the configuration
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	flag.Parse()

	serviceCfg := viper.Sub("service")
	s := server.Server{
		LDAP:    serviceCfg.GetString("ldap"),
		Address: serviceCfg.GetString("address"),
		Proxy:   serviceCfg.GetString("proxy"),
	}

	s.Initialize()
}
