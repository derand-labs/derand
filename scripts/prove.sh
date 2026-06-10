#!/usr/bin/env bash
set -euo pipefail

run_cmd () {
  echo
  echo ">>> $*"
  echo
  "$@"
}

run_cmd ./build/corevdf prove --system $1
run_cmd ./build/corevdf verify --system $1
