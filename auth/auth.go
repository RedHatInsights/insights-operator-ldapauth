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

// Package auth contains implementation of authentication mechanism used to
// validate user credentials and create auth token.
//
// Generated documentation is available at:
// https://godoc.org/github.com/RedHatInsights/insights-operator-ldapauth/auth
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-operator-ldapauth/packages/auth/auth.html
package auth

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/RedHatInsights/insights-operator-utils/responses"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/ldap.v3"
)

// NoAccessMessage - Error message for user with no access
const NoAccessMessage = "User has no rights to access"

// Token JWT claims struct
type Token struct {
	Login string
	jwt.StandardClaims
}

// Account a struct to rep user account
type Account struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// GetTokenPasswordFromEnv tries to read token password from the environment
// variable. Content of the variable (string) is converted into array of bytes
// to be usable.
func GetTokenPasswordFromEnv() []byte {
	return []byte(os.Getenv("token_password"))
}

func createLdapConnection(ldapHost string) (*ldap.Conn, error) {
	var err error
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", ldapHost, 389))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func ldapAuth(login, password string, client ldap.Client) (bool, error) {
	var err error

	// Reconnect with TLS
	// disable "G402 (CWE-295): TLS InsecureSkipVerify set true."
	// #nosec G402
	err = client.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return false, err
	}
	defer client.Close()

	// // Search for the given username
	searchRequest := ldap.NewSearchRequest(
		"dc=redhat,dc=com",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(uid=%s))", login),
		[]string{"dn", "memberOf"},
		nil,
	)

	sr, err := client.Search(searchRequest)
	if err != nil {
		return false, err
	}

	if len(sr.Entries) != 1 {
		return false, errors.New("User does not exist or too many entries returned")
	}

	isInGroup := false

	for i := 0; i < len(sr.Entries[0].Attributes[0].Values); i++ {
		entries := strings.Split(sr.Entries[0].Attributes[0].Values[i], ",")
		if entries[0] == "cn=ccx-dev" {
			isInGroup = true
			break
		}
	}

	if !isInGroup {
		return false, errors.New(NoAccessMessage)
	}

	// Bind as the user to verify their password
	err = client.Bind(sr.Entries[0].DN, password)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Authenticate - validate user credentials and create auth token
func Authenticate(login, password, ldapHost string) (map[string]interface{}, error) {
	var err error
	account := &Account{}
	account.Login = login

	conn, err := createLdapConnection(ldapHost)
	if err != nil {
		log.Println(err)
		r := responses.BuildResponse(err.Error())
		return r, err
	}
	ok, err := ldapAuth(login, password, conn)

	// attempt the authentication
	if err != nil || !ok {
		log.Println(err)
		r := responses.BuildResponse(err.Error())
		return r, err
	}

	// Create JWT token
	tk := &Token{Login: account.Login}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString(GetTokenPasswordFromEnv())

	// Store the token in the response
	account.Token = tokenString

	resp := responses.BuildResponse("ok")
	resp["account"] = account
	return resp, nil
}
