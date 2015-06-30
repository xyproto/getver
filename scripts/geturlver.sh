#!/bin/bash

filename=PKGBUILD
if [[ $1 != '' ]]; then
  filename="$1"
fi
shift 1

verbose=F
if [[ $1 != '' ]]; then
  verbose=T
fi

if [[ ! -f $filename ]]; then
  echo "Can't find $filename"
  exit 1
fi

cd $(dirname "$filename")
filename=$(basename "$filename")

# Use extra parameters, if specified
params=$(grep -a '^# getver:' $filename | sed -n -e 's/^.*getver: //p')

# Retrieve the URL from the file
url=$(grep url= "$filename" | cut -d\" -f2 | cut -d"'" -f2)

# Output the command
if [[ $verbose == T ]]; then
  echo getver $params "$@" "$url"
fi

# Retrieve the latest version number
getver $params "$@" "$url"
