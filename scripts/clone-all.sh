#!/bin/bash
echo Cloning/getting packages used by AusOcean
Repos=(\
  "golang.org/x/sys"\
  "gopkg.in/natefinch/lumberjack.v2"\
  "go.uber.org/zap"\
  "github.com/golang/glog"\
  "github.com/robfig/cron"\
  "github.com/kidoman/embd"\
  "github.com/yobert/alsa"\
  "github.com/adrianmo/go-nmea"\
  "github.com/jacobsa/go-serial"\
  "github.com/tarm/serial"\
  "github.com/Comcast/gots"\
  "github.com/pkg/errors"\
  "github.com/ausocean/utils"\
  "github.com/ausocean/iot"\
  "github.com/ausocean/av"\
  "github.com/ausocean/ocscigo"\
  "github.com/ausocean/iotsvc"\
)
pushd ~
if [ ! -d "go" ]; then
  mkdir go
fi
cd go
if [ ! -d "src" ]; then
  mkdir src
fi
cd src
src=$PWD
for repo in ${Repos[@]}; do
    IFS=/; read cloud account dir <<< "$repo"
    if [[ "$cloud" = go* ]]; then
      echo go get -u "$repo"
      go get -u "$repo"
    else
      if [ ! -d "$src/$cloud/$account" ]; then
        mkdir -p "$src/$cloud/$account"
      fi
      cd "$src/$cloud/$account"
      if [ -d "$dir" ]; then
        echo "$repo" already exists
        continue
      fi
      echo git clone "https://$repo.git"
      git clone "https://$repo.git"
    fi
done
popd
echo Note: The ausocean test and rig repositories have not been cloned.
