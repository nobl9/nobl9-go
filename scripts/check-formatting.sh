#!/usr/bin/env bash

set -e

TMP_DIR=$(mktemp -d)

cp -r . "$TMP_DIR"

make -C "$TMP_DIR" install/yarn
make -C "$TMP_DIR" install/goimports
make -C "$TMP_DIR" format

CHANGED=$(git -C "$TMP_DIR" diff --name-only ${ENUM_PATH})
if [ -n "${CHANGED}" ]; then
  echo >&2 "There are generated code changes that haven't been committed: ${CHANGED}"
  exit 1
else
  echo "Looks good!"
fi
