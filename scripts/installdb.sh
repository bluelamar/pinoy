#!/bin/bash

# install this after the installsvr.sh

#if [ $# -lt 2 ]; then
#  echo "Usage: $0 <db-user-name> <db-user-passwd>"
#  exit 1
#fi

docker pull mongo:4.2.2
DBPATH=/var/opt/mongodb
DBCFG=/var/opt/mongocfg
#DBPATH=/var/opt/couchdb
mkdir -p $DBPATH
chmod 755 $DBPATH
#mkdir -p $DBPATH/db
#chmod 770 $DBPATH/db
#mkdir -p $DBPATH/logs
#chmod 770 $DBPATH/logs

#docker build -t ${CNAME} -f Dockerfile_cdb .

docker run -d -p 27017-27019:27017-27019 -v ${DBPATH}:/data/db -v ${DBCFG}:/data/configdb --name mongodb mongo:4.2.2
#docker update --restart=always <container>
CONTID=`docker ps | grep "mongo:4.2.2" | awk '{ printf $1 }'`
if [ -z "$CONTID" ]; then
  echo "mongodb not running in docker"
  exit 0
fi
docker update --restart=always $CONTID

#/etc/init.d/couchdb start

#echo "Sleeping for 30 seconds to let the db come up before initializing it..."
#sleep 30

# dbusername dbuserpasswd
#./scripts/init_cdb.sh $1 $2

# this is interactive so copy paste the commands from the mongo init scripts
# 0. scripts/admin.mongo  // not checked into github
# 1. scripts/mdb.mongo
# 2: scripts/room_rates.mongo 
# 3: scripts/rooms.mongo 
# 4: scripts/staff.mongo  // not checked into github
docker exec -it mongodb mongo
#docker exec -it -v $PWD:$PWD mongodb mongo 

