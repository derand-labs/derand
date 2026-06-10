#!/usr/bin/env bash
set -euo pipefail

run_cmd () {
  echo
  echo ">>> $*"
  echo
  "$@"
}

run_cmd ./build/zkvdf compile --system $1 --circuit hash_to_form
run_cmd ./build/zkvdf setup   --system $1 --circuit hash_to_form

run_cmd ./build/zkvdf compile --system $1 --circuit intermediate_pow
run_cmd ./build/zkvdf setup   --system $1 --circuit intermediate_pow

if [[ "${2:-false}" = "true" ]]; then
  run_cmd ./build/zkvdf compile --system $1 --circuit rc_verifier
  run_cmd ./build/zkvdf setup   --system $1 --circuit rc_verifier --sol
else
  run_cmd ./build/zkvdf compile --system $1 --circuit rc_verifier_phase_1
  run_cmd ./build/zkvdf setup   --system $1 --circuit rc_verifier_phase_1

  run_cmd ./build/zkvdf compile --system $1 --circuit rc_verifier_phase_2
  run_cmd ./build/zkvdf setup   --system $1 --circuit rc_verifier_phase_2 --sol
fi
