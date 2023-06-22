#!/bin/sh

set -e

make generate

CHANGED=$(git diff --name-only sdk/kind_enum)
if [ -n "${CHANGED}" ]; then
  echo >&2 "There are generated code changes that haven't been committed: ${CHANGED}"
  exit 1
fi
