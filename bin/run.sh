
#!/bin/sh

### functions definition :

killFunc()
{
    strs=`netstat -nap | grep $@ | grep LISTEN | awk {' print $7'}`

    if [ ${#strs} -eq 0 ]; then
        return
    fi
    echo "killing  $@"
    IFS='/' read -r -a array <<< "$strs"
    echo "killing  pid ${array[0]}"
    kill -9 "${array[0]}"
}

startFunc()
{
    cd $exe/$1
    nohup ./$2 &
}

statusFunc()
{
    strs=`netstat -nap | grep tcp_server | grep LISTEN | awk {' print $7'}`
    if [ ${#strs} -eq 0 ]; then
        echo "tcp server not working"
    else
        echo $strs
    fi

    strs=`netstat -nap | grep web_server | grep LISTEN | awk {' print $7'}`
    if [ ${#strs} -eq 0 ]; then
        echo "web server not working"
    else
        echo $strs
    fi
}

usage()
{
    echo "usage:"
    echo "To check status:"
    echo "  sh run.sh status"
    echo "To start web server or / and tcp server:"
    echo "  sh run.sh start web | tcp | all"
    echo "To start web server or / and tcp server:"
    echo "  sh run.sh stop web | tcp | all"
}

### main function :

if [ $# -eq 0 ]; then
    usage
elif [ $# -eq 1 ]; then
    if [ $1 = "status" ];then
        statusFunc
    else
        echo " parameter error"
        usage
    fi
elif [ $# -eq 2 ]; then
    if [ $1 = "start" ];then
        path=`dirname $0`
        curr=`pwd`
        exe=$curr/$path/../pkg

        if [ $2 = "tcp" ];then
            killFunc tcp_server
            startFunc tcp tcp_server
        elif [ $2 = "web" ];then
            killFunc web_server
            startFunc web web_server
        elif [ $2 = "all" ];then
	    killFunc tcp_server
	    killFunc web_server
            killFunc mysqld
            killFunc zimg
            killFunc redis
	    chown -R mysql:mysql /var/lib/mysql /var/run/mysqld 
	    nohup /usr/bin/mysqld_safe  &
            nohup /root/redis/src/redis-server  /root/redis/redis.conf --loglevel debug &

	    cd /root/zimg
	    ./zimg conf/zimg.lua
	    cd $exe

            startFunc tcp tcp_server
            startFunc web web_server
        fi
        sleep 1
        statusFunc
    fi

    if [ $1 = "stop" ];then
        if [ $2 = "tcp" ];then
            killFunc tcp_server
        elif [ $2 = "web" ];then
            killFunc web_server
        elif [ $2 = "all" ];then
            killFunc tcp_server
            killFunc web_server
            killFunc zimg
        fi
        sleep 1
        statusFunc
    fi
elif [ $# -gt 2 ];then
    echo "parameter error"
    usage
fi

