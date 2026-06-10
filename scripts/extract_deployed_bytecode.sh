#!/usr/bin/env bash

set -e

JSON_FILE="$1"
PREFIX="$2"
OUT_FILE="$3"

if [ -z "$JSON_FILE" ] || [ -z "$PREFIX" ] || [ -z "$OUT_FILE" ]; then
  echo "Usage: $0 <forge-json-file> <varPrefix> <out-file.go>"
  exit 1
fi

if ! command -v jq &> /dev/null; then
  echo "jq is required but not installed"
  exit 1
fi

# extract deployed bytecode
BYTECODE=$(jq -r '.deployedBytecode.object' "$JSON_FILE")

if [ "$BYTECODE" == "null" ] || [ -z "$BYTECODE" ]; then
  echo "Cannot find deployedBytecode.object in $JSON_FILE"
  exit 1
fi

cat > "$OUT_FILE" <<EOF
package gen

var ${PREFIX}DeployedBin = "${BYTECODE}"
EOF

echo "Generated: $OUT_FILE"
