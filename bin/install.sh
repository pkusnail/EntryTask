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

#build 
cd ../pkg
go build ../src/tcp_server.go
go build ../src/web_server.go


