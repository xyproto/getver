#!/bin/bash

filename=PKGBUILD
if [[ $1 != '' ]]; then
  filename="$1"
fi

if [[ ! -f $filename ]]; then
  echo "Can't read $filename"
  exit 1
fi

cd $(dirname "$filename")
filename=$(basename "$filename")

# Get the new version, but replace "-" with "_"
newver=$(geturlver "$filename" | sed 's/-/_/g')

# Update the pkgver
[ ! -z $newver ] && setconf "$filename" 'pkgver' "$newver"
