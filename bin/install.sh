#!/bin/sh

path=`dirname $0`
curr=`pwd`
exe=$curr/$path/
cd $exe

#cd $GOPATH/src
#git clone git://github.com/alphazero/Go-Redis.git redis
#cd redis
#go install

go get -u github.com/go-redis/redis
go get github.com/go-sql-driver/mysql
go get github.com/tsliwowicz/go-wrk
go get -u github.com/gorilla/mux
go get golang.org/x/crypto/bcrypt
go get github.com/gorilla/sessions
go get github.com/google/uuid
go get github.com/garyburd/redigo/redis

#build 
cd ../pkg
go build ../src/tcp_server/tcp_server.go
mv tcp_server tcp/
go build ../src/http_server/web_server.go
mv web_server web/
cp -r ../src/http_server/tpl web/

