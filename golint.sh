#!/bin/bash

cd "$(dirname $0)"

if ! [ -x "$(command -v golint)" ]
then
    echo -e "${BLUE}Installing golint${NC}"
    GO111MODULE=off go get golang.org/x/lint/golint 2> /dev/null
fi

if golint `go list ./...` |
    grep -v ALL_CAPS |
    grep .; then
  exit 1
fi
