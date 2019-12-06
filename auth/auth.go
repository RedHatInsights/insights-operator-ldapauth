package auth

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	u "github.com/redhatinsights/insights-operator-ldapauth/utils"
	"gopkg.in/ldap.v3"
)

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

func ldapAuth(login, password, ldapHost string) (bool, error) {
	var err error
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", ldapHost, 389))
	if err != nil {
		return false, err
	}
	defer l.Close()

	// Reconnect with TLS
	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return false, err
	}

	// // Search for the given username
	searchRequest := ldap.NewSearchRequest(
		"dc=redhat,dc=com",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(uid=%s))", login),
		[]string{"dn", "memberOf"},
		nil,
	)

	sr, err := l.Search(searchRequest)
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
		return false, errors.New("User has no rights to access")
	}

	// userdn := sr.Entries[0].DN

	// Bind as the user to verify their password
	err = l.Bind("uid="+login+",ou=users,dc=redhat,dc=com", password)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Authenticate - validate user credentials and create auth token
func Authenticate(login, password, ldap string) (map[string]interface{}, error) {
	var err error
	account := &Account{}
	account.Login = login

	ok, err := ldapAuth(login, password, ldap)

	// attempt the authentication
	if err != nil || !ok {
		log.Println(err)
		r := u.BuildResponse(err.Error())
		return r, err
	}

	//Create JWT token
	tk := &Token{Login: account.Login}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString //Store the token in the response

	resp := u.BuildResponse("ok")
	resp["account"] = account
	return resp, nil
}
