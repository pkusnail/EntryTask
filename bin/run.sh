
#!/bin/sh

path=`dirname $0`
curr=`pwd`
exe=$curr/$path/../pkg

#ps -ef | grep tcp_server | grep -v grep | grep -v tail | awk {'print $2'}  | xargs kill -9
#ps -ef | grep web_server | grep -v grep | grep -v tail | awk {'print $2'}  | xargs kill -9

strs=`netstat -nap | grep tcp_server | grep LISTEN | awk {' print $7'}`
IFS='/' read -r -a array <<< "$strs"
echo "killing tcp_server pid ${array[0]}"
kill -9 "${array[0]}"

strs=`netstat -nap | grep web_server | grep LISTEN | awk {' print $7'}`
IFS='/' read -r -a array <<< "$strs"
echo "killing web_server pid ${array[0]}"
kill -9 "${array[0]}"

cd $exe/tcp
nohup ./tcp_server &
cd $exe/web
nohup ./web_server &

