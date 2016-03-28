#!/bin/bash

maintainer="$1"
if [[ $1 == "" ]]; then
  echo 'Syntax: maintainer [packagedir]'
  exit 1
fi

packagedir="$2"
if [[ $2 == "" ]]; then
  packagedir=~/archpackages/community
fi

cd "$packagedir"
basedir="$(pwd)"
for pkgbuild in $(ag "Maintainer: $maintainer" -G trunk/PKGBUILD -l); do
  dir="$(dirname "$pkgbuild")"
  filename=$(basename "$pkgbuild")
  projectname="$(basename "$(dirname "$dir")")"

  # echo "pkgbuild: $pkgbuild"
  # echo "dir: $dir"
  # echo "filename: $filename"
  #echo "projectname: $projectname"

  cd "$dir"

  # Get the old version
  oldver=$(grep 'pkgver=' "$filename" | head -1 | cut -d'=' -f2)

  # echo "oldver: $oldver"

  # Get the new version, but replace "-" with "_"
  newver=$(geturlver "$filename" | sed 's/-/_/g')

  # echo "newver: $newver"

  # Check if there is a new version
  if [[ $newver != $oldver ]]; then
    echo "There might be a new version of $projectname ($oldver -> $newver)"
  else
    echo "No new version of $projectname ($oldver == $newver)"
  fi

  cd "$basedir"

done


