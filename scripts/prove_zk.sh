#!/usr/bin/env bash
set -euo pipefail

run_cmd () {
  echo
  echo ">>> $*"
  echo
  "$@"
}

# Proving
# run_cmd ./build/zkvdf prove --system $1 --proof $2 --circuit hash_to_form
# run_cmd ./build/zkvdf prove --system $1 --proof $2 --circuit intermediate_pow


# if [ "${3:-false}" = "true" ]; then
#   run_cmd ./build/zkvdf prove --system $1 --proof $2 --circuit rc_verifier
# else
#   run_cmd ./build/zkvdf prove --system $1 --proof $2 --circuit rc_verifier_phase_1
#   run_cmd ./build/zkvdf prove --system $1 --proof $2 --circuit rc_verifier_phase_2
# fi

# Verifying
run_cmd ./build/zkvdf verify --system $1 --proof $2 --circuit hash_to_form
run_cmd ./build/zkvdf verify --system $1 --proof $2 --circuit intermediate_pow

if [ "${3:-false}" = "true" ]; then
  run_cmd ./build/zkvdf verify --system $1 --proof $2 --circuit rc_verifier
else
  run_cmd ./build/zkvdf verify --system $1 --proof $2 --circuit rc_verifier_phase_1
  run_cmd ./build/zkvdf verify --system $1 --proof $2 --circuit rc_verifier_phase_2
fi
