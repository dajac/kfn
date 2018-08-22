#!/usr/bin/env bash

GOPATH=$(go env GOPATH)

PKGS=$(go list ./... | grep -vF /vendor/)
go vet $PKGS
${GOPATH}/bin/golint $PKGS
