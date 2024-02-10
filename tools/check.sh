#!/usr/bin/env nix-shell
#! nix-shell -I nixpkgs=https://github.com/NixOS/nixpkgs/archive/f8e2ebd66d097614d51a56a755450d4ae1632df1.tar.gz
#! nix-shell -p go_1_22
#! nix-shell -i sh --pure
# shellcheck shell=sh

set -eu
cd "$(dirname "$(realpath "$0")")/.."
set -x

go vet ./...
go test -v -parallel "$(grep -c -E "^processor.*[0-9]+" "/proc/cpuinfo")" ./...
