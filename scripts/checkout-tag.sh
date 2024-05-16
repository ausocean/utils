#!/bin/bash
# Check out AusOcean repos for a given Git tag
Usage="Usage: checkout-tag.sh tag"
Repos=(utils iot av)
BaseDir=$GOPATH/src/github.com/ausocean
Tag="$1"
if [ -z "$Tag" ]; then
    echo "$Usage"
    exit 1
fi
if [ -z "$GOPATH" ]; then
    echo "Error: GOPATH not defined"
    exit 1
fi
for repo in ${Repos[@]}; do
  if [ ! -d "$BaseDir/$repo" ]; then
    echo Creating $BaseDir/$repo
    mkdir $BaseDir/$repo
    cd $BaseDir/$repo
    git init
    git remote add origin https://github.com/ausocean/$repo.git
  else
    cd $BaseDir/$repo
  fi
  git fetch --depth=1 origin refs/tags/$Tag:refs/tags/$Tag
  git checkout --force tags/$Tag
done
