#!/usr/bin/env bash

ROOT=${ROOT:-$(git rev-parse --show-toplevel)}

ORG=dajac
NAME=kfn-operator
VERSION=0.0.1
#VERSION=$(git describe --tags --dirty)
#COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null)
#DATE=$(date "+%Y-%m-%d")

if [[ "$(pwd)" != "${ROOT}" ]]; then
  echo "you are not in the root of the repo" 1>&2
  echo "please cd to ${ROOT} before running this script" 1>&2
  exit 1
fi

docker build -t ${ORG}/${NAME}:${VERSION} .
