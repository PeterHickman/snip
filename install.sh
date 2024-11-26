#!/bin/sh

BINARY='/usr/local/bin'
APP=snip

echo "Building $APP"
go build -ldflags="-s -w" $APP.go

echo "Installing $APP to $BINARY"
install $APP $BINARY

echo "Create link ${APP}l for ${APP} --list"
ln -sf $BINARY/$APP $BINARY/snipl

echo "Create link ${APP}d for ${APP} --delete"
ln -sf $BINARY/$APP $BINARY/snipd

echo "Create link ${APP}i for ${APP} --import"
ln -sf $BINARY/$APP $BINARY/snipi

echo "Create link ${APP}e for ${APP} --export"
ln -sf $BINARY/$APP $BINARY/snipe

echo "Create link ${APP}s for ${APP} --search"
ln -sf $BINARY/$APP $BINARY/snips

echo "Removing the build"
rm $APP