
#build stage
FROM golang:1.10.3-alpine AS builder
RUN apk add -U --no-cache ca-certificates git bash
WORKDIR /go/src/github.com/dajac/kfn/
COPY . .
RUN go get -u github.com/golang/dep/cmd/dep && dep ensure -v
RUN ROOT=$(pwd) ./hack/build.bash

#final stage
FROM alpine:latest
COPY --from=builder /go/src/github.com/dajac/kfn/bin/kfn-operator /kfn-operator
ENTRYPOINT ./kfn-operator
