# Insights operator LDAP Auth

[![Go Report Card](https://goreportcard.com/badge/github.com/RedHatInsights/insights-operator-ldapauth)](https://goreportcard.com/report/github.com/RedHatInsights/insights-operator-ldapauth)[![Build Status](https://travis-ci.org/RedHatInsights/insights-operator-ldapauth.svg?branch=master)](https://travis-ci.org/RedHatInsights/insights-operator-ldapauth)

## Overview

The whole system consists of several components: CLI client, LDAP Auth service, and the Insights operator instrumentation service:

![Components](doc/components.png)

### Authorization part

![Diagram2](doc/auth_sequence_002.png)

### Authorization sequence diagram for known user with proper token

![Diagram7](doc/auth_sequence_007.png)

### Authorization sequence diagram in case of any error

![Diagram8](doc/auth_sequence_008.png)

## Starting

By default application starting on port `8081`, but it can be changed in configuration file `config.toml`.

```Bash
go build # Build application
./insights-operator-ldapauth # Start application
```

## Authentication

For authentication is used POST request to `/api/v1/login` with credentials:
```JSON
{
	"login": "your-ldap-login",
	"password": "your-ldap-password"
}
```

For now it connecting directly to RedHat LDAP, so for running this application correctly you should be connected to RedHat VPN. After you recieve `token`, you can use it in requests as `Bearer Token`.

## RestAPI

Application has only one route is `/api/v1/login`, requests to other routes will be proxied to `insights-operator-controller`.
