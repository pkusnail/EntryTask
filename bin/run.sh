
#!/bin/sh

path=`dirname $0`
curr=`pwd`
exe=$curr/$path/../pkg

ps -ef | grep tcp_server | grep -v grep | grep -v tail | awk {'print $2'}  | xargs kill -9
ps -ef | grep web_server | grep -v grep | grep -v tail | awk {'print $2'}  | xargs kill -9
cd $exe/tcp
nohup ./tcp_server &
cd $exe/web
nohup ./web_server &

