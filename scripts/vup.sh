#!/bin/bash

filename=PKGBUILD
if [[ $1 != '' ]]; then
  filename="$1"
fi

if [[ ! -f $filename ]]; then
  echo "Can't find $filename"
  exit 1
fi

cd $(dirname "$filename")
filename=$(basename "$filename")

# Update the pkgver
bumpver "$filename"

# Update the hash sums
updpkgsums "$filename"
