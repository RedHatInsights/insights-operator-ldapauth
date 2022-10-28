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
// https://redhatinsights.github.io/insights-operator-ldapauth/packages/server/auth.html

import (
	"context"
	"encoding/json"
	"github.com/RedHatInsights/insights-operator-utils/responses"
	jwt "github.com/dgrijalva/jwt-go"
	auth "github.com/redhatinsights/insights-operator-ldapauth/auth"
	"log"
	"net/http"
	"strings"
)

type contextKey string

const (
	contextKeyUser = contextKey("user")
)

// Login handler for login route
func (s *Server) Login(writer http.ResponseWriter, request *http.Request) {
	account := &auth.Account{}

	// decode the request body into struct and failed if any error occur
	err := json.NewDecoder(request.Body).Decode(account)
	if err != nil {
		err := responses.SendBadRequest(writer, "Invalid request")
		if err != nil {
			log.Println(err)
		}
		return
	}
	resp, err := auth.Authenticate(account.Login, account.Password, s.LDAP)
	if err != nil {
		err := responses.Send(http.StatusUnauthorized, writer, resp)
		if err != nil {
			log.Println(err)
		}
		return
	}
	err = responses.SendOK(writer, resp)
	if err != nil {
		log.Println(err)
	}
}

// JWTAuthentication - middleware for authenticate user by Token
func (s *Server) JWTAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// List of endpoints that doesn't require auth
		notAuth := []string{APIPrefix + "login"}

		// current request path
		requestPath := r.URL.Path

		// check if request does not need authentication, serve the request if it doesn't need it
		for _, value := range notAuth {

			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		// grab the token from the header
		tokenHeader := r.Header.Get("Authorization")

		if tokenHeader == "" {
			// Token is missing, return with HTTP error code 403 Unauthorized
			err := responses.SendForbidden(w, "Missing auth token")
			if err != nil {
				log.Println(err)
			}
			return
		}

		// the token normally comes in format `Bearer {token-body}`, we
		// check if the retrieved token matched this requirement
		splitted := strings.Split(tokenHeader, " ")
		if len(splitted) != 2 {
			err := responses.SendForbidden(w, "Invalid/Malformed auth token")
			if err != nil {
				log.Println(err)
			}
			return
		}

		// grab the token part, what we are truly interested in
		tokenPart := splitted[1]
		tk := &auth.Token{}

		token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
			return auth.GetTokenPasswordFromEnv(), nil
		})

		if err != nil {
			// malformed token detected, returns with HTTP code 403 as usual
			err := responses.SendForbidden(w, "Malformed authentication token")
			if err != nil {
				log.Println(err)
			}
			return
		}

		if !token.Valid {
			// token is invalid, maybe not signed on this server
			err := responses.SendForbidden(w, "Token is not valid.")
			if err != nil {
				log.Println(err)
			}
			return
		}

		// everything went well, proceed with the request and set the
		// caller to the user retrieved from the parsed token
		ctx := context.WithValue(r.Context(), contextKeyUser, tk.Login)
		r = r.WithContext(ctx)
		// Proceed to proxy
		next.ServeHTTP(w, r)
	})
}
