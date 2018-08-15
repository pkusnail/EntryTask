
#!/bin/sh

path=`dirname $0`
curr=`pwd`
cd $curr/$path/../pkg

ps -ef | grep tcp_server | grep -v grep | awk {'print $2'}  | xargs kill -9
ps -ef | grep web_server | grep -v grep | awk {'print $2'}  | xargs kill -9

nohup ./tcp_server &
nohup ./web_server &

