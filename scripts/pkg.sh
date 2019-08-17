#!/bin/bash

tar cvf pinoy.tar pinoy static scripts config.json cdb-local.ini Dockerfile_cdb

gzip pinoy.tar

