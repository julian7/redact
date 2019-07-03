#!/usr/bin/env bash
cd ${0%/*}

exe=${PWD##*/}
version=$(git describe --tags 2>/dev/null)
: ${version:=untagged}
ldflags='-s -w -X "main.version='$version'"'
if which upx >/dev/null 2>/dev/null; then
    UPX=1
fi

mkdir -p target

build_in() {
    local os="$1"
    local target="$2"
    : ${target:=target/$exe-$os-amd64-$version}
    echo "---> Building for $os"
    GOOS=$os GOARCH=amd64 \
        go build -ldflags "$ldflags" \
        -o "$target" ./cmd/$exe || exit 1
    if [[ "$UPX" == 1 ]]; then
        upx "$target"
    fi
}

if (( $# == 1 )); then
    build_in $1 $exe
else
    for os in windows linux darwin; do
        build_in $os
    done
fi
