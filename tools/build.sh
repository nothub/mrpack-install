#!/usr/bin/env nix-shell
#! nix-shell -I nixpkgs=https://github.com/NixOS/nixpkgs/archive/f8e2ebd66d097614d51a56a755450d4ae1632df1.tar.gz
#! nix-shell -p go_1_22 git upx
#! nix-shell -i sh --pure
# shellcheck shell=sh

set -eu
cd "$(dirname "$(realpath "$0")")/.."
set -x

# clean up old artifacts
rm -rf "dist"
mkdir -p "dist"

# check for version tag on HEAD commit
tag="$(git tag --points-at HEAD |
    grep -oP 'v[0-9]+\.[0-9]+\.[0-9]+' || true | tail -n 1)"

# is there a version tag pointing to HEAD?
if test -n "${tag}"; then
    # yes, this is a release version
    release="true"
else
    # no, not a release version...
    release="false"
    # get last version tag in history
    tag="$(git describe --tags --abbrev=0 --match v[0-9]* 2>/dev/null |
        grep -oP 'v[0-9]+\.[0-9]+\.[0-9]+')"
    # mark as dev build
    tag="${tag}-indev"
fi

build() (

  file="dist/mrpack-install_${1}_${2}"
  if test "$1" = "windows"; then
    file="${file}.exe"
  fi

  # build static binary
  GOOS="$1" GOARCH="$2" go build \
    -tags netgo,timetzdata \
    -ldflags="-X 'hub.lol/mrpack-install/buildinfo.Tag=${tag}' -X 'hub.lol/mrpack-install/buildinfo.Release=${release}' -extldflags=-static" \
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
