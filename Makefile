SHELL := /bin/bash
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin

default: build validate test

get-deps:
	go get -u golang.org/x/lint/golint github.com/golang/dep/cmd/dep

dep:
	$(GOBIN)/dep ensure

codegen:
	./hack/update-codegen.bash

build:
	go fmt ./...
	./hack/build.bash

validate:
	$(GOBIN)/dep check
	./hack/lint.bash

test: build
	./hack/test.bash

clean:
	rm -rf bin/
	rm -rf vendor/

.PHONY: build validate test
