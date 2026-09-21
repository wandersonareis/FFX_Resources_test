#!/bin/bash
# Roda todos os testes Go. Uso: bash scripts/go-test.sh
set -u
export PATH="$HOME/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
cd "$(dirname "$0")/.."
/usr/local/go/bin/go test ./... 2>&1 | grep -Ev "^(ok|\?|---)" | tail -5
echo "---- summary ----"
/usr/local/go/bin/go test ./... 2>&1 | grep -E "^(ok|FAIL|\?)" | sort | uniq -c | sort -rn | head -30
