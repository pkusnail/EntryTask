cd $GOPATH/src
git clone git://github.com/alphazero/Go-Redis.git redis
cd redis
go install

go get -u github.com/gorilla/mux
go get golang.org/x/crypto/bcrypt
go get github.com/gorilla/sessions
go get github.com/google/uuid

