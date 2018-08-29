#!/usr/bin/env bash

ROOT=${ROOT:-$(git rev-parse --show-toplevel)}

ORG=dajac
REPO=kfn
NAME=kfn-operator
FULLVERSION=$(git describe --tags --dirty)
VERSION=${FULLVERSION#"v"}
#COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null)
#DATE=$(date "+%Y-%m-%d")
TARGET_MANIFEST=${ROOT}/config/kfn-${VERSION}.yaml

if [[ "$(pwd)" != "${ROOT}" ]]; then
  echo "you are not in the root of the repo" 1>&2
  echo "please cd to ${ROOT} before running this script" 1>&2
  exit 1
fi

docker build -t ${ORG}/${NAME}:${VERSION} .
docker push ${ORG}/${NAME}:${VERSION}

rm ${TARGET_MANIFEST} 2>/dev/null

for file in `ls ${ROOT}/config/`; do
    cat ${ROOT}/config/$file >> ${TARGET_MANIFEST}
    echo "---" >> ${TARGET_MANIFEST}
done;

if [[ -z "$ACCESS_TOKEN" ]]; then
  echo "Unable to release: Github Token not specified" > /dev/stderr
  exit 1
fi

payload=$(
  jq --null-input \
     --arg tag "$VERSION" \
     --arg name "$VERSION" \
     '{ tag_name: $tag, name: $name, draft: true }'
)

response=$(
  curl --fail \
       --header "Authorization: token $ACCESS_TOKEN" \
       --silent \
       --location \
       --data "$payload" \
       "https://api.github.com/repos/${ORG}/${REPO}/releases"
)

upload_url="$(echo "$response" | jq -r .upload_url | sed -e "s/{?name,label}//")"

curl --header "Authorization: token $ACCESS_TOKEN" \
     --header "Content-Type:text/yaml" \
     --data-binary "@${TARGET_MANIFEST}" \
     "$upload_url?name=$(basename "${TARGET_MANIFEST}")"

rm ${TARGET_MANIFEST} 2>/dev/null
