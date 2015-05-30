#!/bin/sh
if [[ `basename "$@"` != PKGBUILD ]]; then
  echo 'First param should be a PKGBUILD file'
  exit 1
fi
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

# Update pkgrel
setconf "$filename" pkgrel+=1
