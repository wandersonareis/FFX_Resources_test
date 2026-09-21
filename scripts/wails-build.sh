#!/bin/bash
# Build Wails (Linux) com tag webkit2_41. Uso: bash scripts/wails-build.sh
set -u
export PATH="$HOME/.bun/bin:$HOME/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
cd "$(dirname "$0")/.."
wails build -tags webkit2_41 2>&1 | tail -30
