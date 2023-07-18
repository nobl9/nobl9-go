#!/bin/sh

set -e

make generate

ENUM_PATH="*_enum.go"

CHANGED=$(git diff --name-only "${ENUM_PATH}")
if [ -n "${CHANGED}" ]; then
  echo >&2 "There are generated code changes that haven't been committed: ${CHANGED}"
  git restore "${ENUM_PATH}"
  exit 1
else
  echo "Looks good!"
fi
