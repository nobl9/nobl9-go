#!/usr/bin/env bash

set -e

ENUM_PATH=sdk/kind_enum.go
TMP_DIR=$(mktemp -d)

cleanup_git() {
  git -C "$TMP_DIR" clean -df
  git -C "$TMP_DIR" checkout -- .
}

main() {
  cp -r . "$TMP_DIR"
  cleanup_git
  
  make -C "$TMP_DIR" generate
  
  CHANGED=$(git -C "$TMP_DIR" diff --name-only ${ENUM_PATH})
  if [ -n "${CHANGED}" ]; then
    printf >&2 "There are generated code changes that are not committed:\n%s\n" "$CHANGED"
    exit 1
  else
    echo "Looks good!"
  fi
}

main "$@"
