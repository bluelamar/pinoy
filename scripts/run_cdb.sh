
CNAME=pinoy-cdb-1

# docker run -it -v $PWD:$PWD --security-opt seccomp-unconfined $CNAME /bin/bash

#docker run --rm -v $PWD:$PWD -w "${PWD}" "${CNAME}" sh scripts/run_cdb.sh ${PWD}

# run the db
#sudo docker run -d -v ~/couchdb:/usr/local/var/lib/couchdb $CNAME

#sudo docker run -d -p 5984:5984 -v ~/couchdb:/usr/local/var/lib/couchdb $CNAME

if [ ! -d $(pwd)/cdb ]; then
  mkdir $(pwd)/cdb
fi
#sudo docker run -p 5984:5984 -p 5986:5986 -v $(pwd)/cdb:/opt/couchdb/data $CNAME
sudo docker run -p 5984:5984 -p 5986:5986 -v $(pwd)/cdb:/var/opt/couchdb $CNAME


#if [ ! -d /var/opt/couchdb ]; then
#  sudo mkdir /var/opt/couchdb
#  sudo chmod 777 /var/opt/couchdb
#fi

#sudo docker run -p 5984:5984 -p 5986:5986 -v /var/opt/couchdb:/var/opt/couchdb $CNAME

