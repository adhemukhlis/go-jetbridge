#!/bin/bash

# proto-export.sh
# This script exports internal protos to the contract directory and strips
# language-specific options to ensure a "Pure Universal Contract".

set -euo pipefail

CONTRACT_DIR="./contract"

echo "📦 Exporting protos to ${CONTRACT_DIR}..."

# 1. Perform buf export
# Note: This will copy protos and their dependencies to the contract folder.
buf export . --output "${CONTRACT_DIR}"

echo "🧹 Cleaning up language-specific options for universal consumption..."

# 2. Iterate through proto files in the contract directory and its subdirectories
# We use a glob for simplicity as requested (no 'find')
# Note: this assumes a relatively flat or known structure, or we use shell globstar
shopt -s globstar 2>/dev/null || true

for file in "${CONTRACT_DIR}"/**/*.proto "${CONTRACT_DIR}"/*.proto; do
    # Check if file exists to avoid errors with globs
    if [ -f "$file" ]; then
        echo "  - Purifying $(basename "$file")"
        
        # Remove lines containing 'option go_package'
        # Using a temporary file for compatibility across macOS and Linux
        sed '/option go_package/d' "$file" > "$file.tmp" && mv "$file.tmp" "$file"
    fi
done

echo "✅ Pure Universal Contracts are ready in ${CONTRACT_DIR}"
