#!/usr/bin/env bash

ROOT=${ROOT:-$(git rev-parse --show-toplevel)}

NAME=kfn-operator

PLATFORM=$(go env GOOS)
ARCH=$(go env GOARCH)
GOPATH=$(go env GOPATH)

if [[ "$(pwd)" != "${ROOT}" ]]; then
  echo "you are not in the root of the repo" 1>&2
  echo "please cd to ${ROOT} before running this script" 1>&2
  exit 1
fi

GO_BUILD_CMD="go build"
GO_BUILD_LDFLAGS=""

mkdir -p "${ROOT}/bin"

echo "Building for ${PLATFORM}/${ARCH}"
GOARCH=${ARCH} GOOS=${PLATFORM} ${GO_BUILD_CMD} -ldflags "${GO_BUILD_LDFLAGS}" \
    -o "${ROOT}/bin/${NAME}" ./cmd/${NAME}/
