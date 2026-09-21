#!/bin/bash
# Sobe o dev-server brevemente para validar o builder. Uso: bash scripts/ng-serve-check.sh
set -u
export PATH="$HOME/.bun/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
cd "$(dirname "$0")/../frontend"
"$HOME/.bun/bin/bunx" ng serve --port 4400 > /tmp/ngserve.log 2>&1 &
SERVER_PID=$!
sleep 40
kill "$SERVER_PID" 2>/dev/null
grep -aEi "deprecat|error|compiled|complete|localhost:4400" /tmp/ngserve.log | head -12
