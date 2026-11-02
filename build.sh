#!/bin/sh
set -e

cd "$(dirname "$0")"

name=psx-helper
targets="windows/amd64 windows/386 linux/amd64 linux/386 darwin/amd64"

rm -rf out
mkdir -p out

for t in $targets; do
	os=${t%/*}
	arch=${t#*/}
	ext=
	[ "$os" = windows ] && ext=.exe
	output="out/${name}-${os}-${arch}${ext}"
	echo "building $output"
	CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags="-s -w" -o "$output" .
done
