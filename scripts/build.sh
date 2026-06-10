#!/usr/bin/env bash
set -euo pipefail

run_cmd () {
  echo
  echo ">>> $*"
  echo
  "$@"
}

VERSIONS=(
  "16"
  "16_2"
  "1024"
  "2048"
  "3072"
  "4096"
  "6656"
)

get_config () {
  local ver="$1"
  case "$ver" in
    "16")
      LIMB_BITS="16"
      SPLIT_EXP="1"
      ;;
    "16_2")
      LIMB_BITS="16"
      SPLIT_EXP="2"
      ;;
    "1024")
      LIMB_BITS="114"
      SPLIT_EXP="1"
      ;;
    "2048")
      LIMB_BITS="121"
      SPLIT_EXP="2"
      ;;
    "3072")
      LIMB_BITS="123"
      SPLIT_EXP="4"
      ;;
    "4096")
      LIMB_BITS="121"
      SPLIT_EXP="4"
      ;;
    "6656")
      LIMB_BITS="122"
      SPLIT_EXP="8"
      ;;
    *)
      exit 1
      ;;
  esac
}

run_version () {
  local VERSION="$1"
  get_config "$VERSION"
  
  if [[ "$VERSION" == "16" || "$VERSION" == "16_2" ]]; then
    run_cmd ./build/corevdf setup \
      --d-bits 16 \
      --limb-bits "$LIMB_BITS" \
      --split-exp "$SPLIT_EXP" \
      --hash-to-form-nb-generators 8 \
      --hash-to-form-steps 2
  else
    run_cmd ./build/corevdf setup \
      --d-bits ${VERSION} \
      --limb-bits "$LIMB_BITS" \
      --split-exp "$SPLIT_EXP" \
      --hash-to-form-nb-generators 9097 \
      --hash-to-form-steps 26
  fi
}

FOUND=0
for v in "${VERSIONS[@]}"; do
  if [[ "$v" == "$1" ]]; then
    FOUND=1
    break
  fi
done

if [[ "$FOUND" -eq 0 ]]; then
  echo "Invalid version: $1"
  echo "Valid versions: ${VERSIONS[*]}"
  exit 1
fi

run_version "$1"
