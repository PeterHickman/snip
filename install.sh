#!/bin/sh

make_link() {
  echo "Create link ${APP}$1 for ${APP} $2"
  ln -sf $BINARY/$APP $BINARY/${APP}$1
}

BINARY='/usr/local/bin'
APP=snip

echo "Building $APP"
go build -ldflags="-s -w" $APP.go

echo "Installing $APP to $BINARY"
install $APP $BINARY

make_link l --list
make_link d --delete
make_link i --import
make_link e --export
make_link s --search

echo "Removing the build"
rm $APP