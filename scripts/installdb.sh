#!/bin/bash

# install this after the installsvr.sh

if [ $# -lt 2 ]; then
  echo "Usage: $0 <db-user-name> <db-user-passwd>"
  exit 1
fi

DBPATH=/var/opt/couchdb
mkdir -p $DBPATH
chmod 755 $DBPATH
mkdir -p $DBPATH/db
chmod 770 $DBPATH/db
mkdir -p $DBPATH/logs
chmod 770 $DBPATH/logs

docker build -t ${CNAME} -f Dockerfile_cdb .

/etc/init.d/couchdb start

echo "Sleeping for 30 seconds to let the db come up before initializing it..."
sleep 30

# dbusername dbuserpasswd
./scripts/init_cdb.sh $1 $2

