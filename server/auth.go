package server

import (
	"context"
	"encoding/json"
	jwt "github.com/dgrijalva/jwt-go"
	auth "github.com/redhatinsights/insights-operator-ldapauth/auth"
	u "github.com/redhatinsights/insights-operator-ldapauth/utils"
	"net/http"
	"os"
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
		status := u.BuildResponse("Invalid request")
		u.SendError(writer, status)
		return
	}
	resp, err := auth.Authenticate(account.Login, account.Password, s.LDAP)
	if err != nil {
		u.SendUnauthorized(writer, resp)
		return
	}
	u.SendResponse(writer, resp)
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

		response := make(map[string]interface{})
		tokenHeader := r.Header.Get("Authorization") //Grab the token from the header

		if tokenHeader == "" { //Token is missing, returns with error code 403 Unauthorized
			response = u.BuildResponse("Missing auth token")
			u.SendForbidden(w, response)
			return
		}

		splitted := strings.Split(tokenHeader, " ") //The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
		if len(splitted) != 2 {
			response = u.BuildResponse("Invalid/Malformed auth token")
			u.SendForbidden(w, response)
			return
		}

		tokenPart := splitted[1] //Grab the token part, what we are truly interested in
		tk := &auth.Token{}

		token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("token_password")), nil
		})

		if err != nil { //Malformed token, returns with http code 403 as usual
			response = u.BuildResponse("Malformed authentication token")
			u.SendForbidden(w, response)
			return
		}

		if !token.Valid { //Token is invalid, maybe not signed on this server
			response = u.BuildResponse("Token is not valid.")
			u.SendForbidden(w, response)
			return
		}

		//Everything went well, proceed with the request and set the caller to the user retrieved from the parsed token
		ctx := context.WithValue(r.Context(), contextKeyUser, tk.Login)
		r = r.WithContext(ctx)
		// Proceed to proxy
		next.ServeHTTP(w, r)
	})
}
