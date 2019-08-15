#!/bin/bash

# run this script before the database install script

tar xzvf pinoy.tar.gz

cp scripts/couchdb.deploy /etc/init.d/couchdb
chmod 755 /etc/init.d/couchdb
cp scripts/pinoy.deploy /etc/init.d/pinoy
chmod 755 /etc/init.d/pinoy

pushd /etc/rc5.d
ln -s ../init.d/couchdb S03couchdb
chmod 777 S03couchdb
ln -s ../init.d/pinoy S04pinoy
chmod 777 S04pinoy
cd ../rc6.d
ln -s ../init.d/couchdb K04couchdb
chmod 777 K04couchdb
ln -s ../init.d/pinoy K03pinoy
chmod 777 K03pinoy
popd

mkdir /etc/pinoy
chmod 755 /etc/pinoy
cp config.json /etc/pinoy/
cp -r static /etc/pinoy/
chmod -R 644 /etc/pinoy/

cp pinoy /usr/bin/
chmod 755 /usr/bin/pinoy

echo "Make directory for log files: ensure config.json is set with this path:"
mkdir /var/log/pinoy
chmod 755 /var/log/pinoy/

