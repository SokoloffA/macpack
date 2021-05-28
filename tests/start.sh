#/bin/bash

set -e

(cd .. && go build)
go build 
./tests $@
