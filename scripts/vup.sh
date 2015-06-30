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

# Download the files
makepkg -g "$filename" 2>/dev/null

# Get the new hash sum
key=$(makepkg -g "$filename" 2>/dev/null | head -1 | cut -d"=" -f1)
value=$(makepkg -g "$filename" 2>/dev/null | cut -d"=" -f2)

# Update the hash sum
if [[ $key ]] && [[ $value ]]; then
  setconf "$filename" "$key" "$value" ')'
fi

