
CNAME=pinoy-cdb-1

# docker run -it -v $PWD:$PWD --security-opt seccomp-unconfined $CNAME /bin/bash

sudo docker build -t ${CNAME} -f Dockerfile_cdb .

#docker run --rm -v $PWD:$PWD -w "${PWD}" "${CNAME}" sh scripts/run_cdb.sh ${PWD}

# run the db
#docker run -d -p 5984:5984 -v ~/couchdb:/usr/local/var/lib/couchdb $CNAME
#docker run -d -v ~/couchdb:/usr/local/var/lib/couchdb $CNAME

