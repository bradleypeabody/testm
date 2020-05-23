#!/bin/bash

set -e

GOOS=linux GOARCH=amd64 go build -o testm .

docker build -t testm .

# docker build -t vugu/testm .
# docker push vugu/testm

# docker run --rm -ti --network=networkname_default -p 8812:8812 testm /bin/testm --dbconn='root:rootpw@tcp4(mysql:3306)/testm?collation=utf8mb4_unicode_ci'
