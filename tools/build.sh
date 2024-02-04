#!/usr/bin/env nix-shell
#! nix-shell -I nixpkgs=https://github.com/NixOS/nixpkgs/archive/e92b6015881907e698782c77641aa49298330223.tar.gz
#! nix-shell -p go_1_21 git upx
#! nix-shell -i sh --pure
# shellcheck shell=sh

set -eu
cd "$(dirname "$(realpath "$0")")/.."
set -x

# clean up old artifacts
rm -rf "dist"
mkdir -p "dist"

# check for version tag on HEAD commit
tag="$(git describe --exact-match "$(git rev-parse HEAD)" 2>/dev/null || true |
  grep -oP 'v[0-9]+\.[0-9]+\.[0-9]+' || true)"

# if HEAD has no version tag
if test -z "${tag}"; then
  # get last version tag in history
  tag="$(git describe --tags --abbrev=0 --match v[0-9]* 2>/dev/null |
    grep -oP 'v[0-9]+\.[0-9]+\.[0-9]+')-indev"
fi

build() (

  file="dist/mrpack-install_${1}_${2}"
  if test "$1" = "windows"; then
    file="${file}.exe"
  fi

  # build static binary
  GOOS="$1" GOARCH="$2" go build \
    -tags netgo,timetzdata \
    -ldflags="-X 'hub.lol/mrpack-install/buildinfo.Tag=${tag}' -extldflags=-static" \
    -o "${file}" \
    .

  chmod +x "${file}"

  # compress with upx
  # ( except for mac because upx mac support requires a feature flag )
  if test "$1" != "darwin"; then
    upx --best --lzma \
      --no-color \
      --no-progress \
      --no-time \
      "${file}"
  fi

)

build linux amd64
build linux arm64
build darwin amd64
build darwin arm64
build windows amd64
