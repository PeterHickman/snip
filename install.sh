#!/bin/sh

BINARY='/usr/local/bin'

echo "Building snip"
go build snip.go

echo "Installing snip to $BINARY"
install -v snip $BINARY

echo "Removing the build"
rm snip
