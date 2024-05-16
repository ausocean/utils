#!/bin/bash
# Reset Git tags to v1.0 and run v1.0 of test client.
AUSOCEAN=$HOME/go/src/github.com/ausocean
if [ ! -f /etc/netsender.conf ]; then
  echo /etc/netsender.conf does not exist
fi
if [ ! -d /var/log/netsender ]; then
  echo /var/log/netsender does not exist
  exit 1
fi
if [ ! -d /var/netsender ]; then
  echo /var/netsender does not exist
  exit 1
fi
if [ -f /var/log/netsender/netsender.log ]; then
  rm /var/log/netsender/netsender.log
  touch /var/log/netsender/netsender.log
fi
if [ -f /var/netsender/tags.conf ]; then
  rm /var/netsender/tags.conf
fi
set -x
cd $AUSOCEAN
if [ -d test ]; then
  rm -rf test
fi    
mkdir test
cd test
git init --quiet
git remote add origin https://github.com/ausocean/test.git
git fetch --quiet origin refs/tags/v1.0:refs/tags/v1.0
git checkout --quiet tags/v1.0
cd test-netsender
make
export PATH=$PATH:$AUSOCEAN/test/test-netsender
test-netsender 
