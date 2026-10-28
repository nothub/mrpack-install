#!/usr/bin/env sh

set -eu
cd "$(dirname "$(realpath "$0")")/.."

echo "#!/bin/sh
set -eu
go run ./tools/readme > README.md
git add README.md
" >.git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
