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

# Get the old version
oldver=$(grep 'pkgver=' "$filename" | head -1 | cut -d'=' -f2)

# Get the new version, but replace "-" with "_"
newver=$(geturlver "$filename" | sed 's/-/_/g')

# Check if there are enough version results
if [[ $newver == Not* ]]; then
  echo "$newver"
  exit 1
fi

# Check if there is a new version
if [[ $newver != $oldver ]]; then
  # Update the pkgver
  [ ! -z $newver ] && (echo "$newver"; setconf "$filename" 'pkgver' "$newver"; setconf "$filename" 'pkgrel' '1')
fi
