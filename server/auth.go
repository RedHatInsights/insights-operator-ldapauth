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

import (
	"context"
	"encoding/json"
	jwt "github.com/dgrijalva/jwt-go"
	auth "github.com/redhatinsights/insights-operator-ldapauth/auth"
	"github.com/redhatinsights/insights-operator-utils/responses"
	"net/http"
	"strings"
)

type contextKey string

const (
	contextKeyUser = contextKey("user")
)

// Login handler for login route
func (s Server) Login(writer http.ResponseWriter, request *http.Request) {
	account := &auth.Account{}
	err := json.NewDecoder(request.Body).Decode(account) //decode the request body into struct and failed if any error occur
	if err != nil {
		responses.SendError(writer, "Invalid request")
		return
	}
	resp, err := auth.Authenticate(account.Login, account.Password, s.LDAP)
	if err != nil {
		responses.SendUnauthorized(writer, resp)
		return
	}
	responses.SendResponse(writer, resp)
}

// JWTAuthentication - middleware for authenticate user by Token
func (s Server) JWTAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		notAuth := []string{APIPrefix + "login"} //List of endpoints that doesn't require auth
		requestPath := r.URL.Path                //current request path

		//check if request does not need authentication, serve the request if it doesn't need it
		for _, value := range notAuth {

			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		tokenHeader := r.Header.Get("Authorization") //Grab the token from the header

		if tokenHeader == "" { //Token is missing, returns with error code 403 Unauthorized
			responses.SendForbidden(w, "Missing auth token")
			return
		}

		splitted := strings.Split(tokenHeader, " ") //The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
		if len(splitted) != 2 {
			responses.SendForbidden(w, "Invalid/Malformed auth token")
			return
		}

		tokenPart := splitted[1] //Grab the token part, what we are truly interested in
		tk := &auth.Token{}

		token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
			return auth.GetTokenPasswordFromEnv(), nil
		})

		if err != nil { //Malformed token, returns with http code 403 as usual
			responses.SendForbidden(w, "Malformed authentication token")
			return
		}

		if !token.Valid { //Token is invalid, maybe not signed on this server
			responses.SendForbidden(w, "Token is not valid.")
			return
		}

		//Everything went well, proceed with the request and set the caller to the user retrieved from the parsed token
		ctx := context.WithValue(r.Context(), contextKeyUser, tk.Login)
		r = r.WithContext(ctx)
		// Proceed to proxy
		next.ServeHTTP(w, r)
	})
}
