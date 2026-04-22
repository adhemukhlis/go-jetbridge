#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${1:-${SCRIPT_DIR}/.env}"

# Load environment variables from .env
if [ -f "$ENV_FILE" ]; then
	echo "📄 Loading environment from $ENV_FILE"
	set -a
	source "$ENV_FILE"
	set +a
else
	echo "❌ Error: .env file not found at $ENV_FILE"
	exit 1
fi

# Required environment variables check
required_vars=("PGHOST" "PGPORT" "PGUSER" "PGDATABASE")
for var in "${required_vars[@]}"; do
	if [ -z "${!var}" ]; then
		echo "❌ Error: $var is not set in .env file"
		exit 1
	fi
done

# Generate to a temporary directory to avoid using dynamic db name as folder name
TEMP_GEN_DIR=$(mktemp -d)
jet -source=PostgreSQL -path="$TEMP_GEN_DIR" -ignore-tables=_prisma_migrations -host=$PGHOST -port=$PGPORT -user=$PGUSER -password=$PGPASSWORD -dbname=$PGDATABASE

# Clean up existing jet gen directory
rm -rf ./gen/jet
mkdir -p ./gen/jet

# Move generated content to gen/jet
# Jet creates a folder with the db name, we move its contents to gen/jet
mv "$TEMP_GEN_DIR/$PGDATABASE/"* ./gen/jet/

# Clean up temp directory
rm -rf "$TEMP_GEN_DIR"

echo "✅ Jet code generated to gen/jet"