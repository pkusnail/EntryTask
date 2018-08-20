package util

import (
	"time"
	"os/exec"
	"hash/fnv"
	"strconv"
	//"redis"
	"log"
	"bufio"
	"strings"
	"os"
	"io"
	"database/sql"
	"sync"
	"github.com/gomodule/redigo/redis"
)

func B2S(bs []uint8) string {
	b := make([]byte, len(bs))
	for i, v := range bs {
		b[i] = byte(v)
	}
	return string(b)
}

func UUID() string {
    out, err := exec.Command("uuidgen").Output()
    if err != nil {
        log.Println(err)
    }
    return strings.Replace(string(out), "\n", "", -1)
}

func Hash(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return strconv.FormatUint(h.Sum64(), 10)
}

type RedisPool struct {
	Pool *redis.Pool;
	Addr *string;
	MaxIdle *int;
	MaxActive *int;
	//IdleTimeout *int;
}

func (rp *RedisPool) NewPool() {
	rp.Pool = &redis.Pool{
		MaxIdle : *rp.MaxIdle,
		MaxActive : *rp.MaxActive,
		IdleTimeout: 5 * time.Second,
		Dial: func () (redis.Conn, error) { return redis.Dial("tcp", *rp.Addr) },
	}
}

func (rp *RedisPool)  RedisSet( key string, val string)  {
    startTime := time.Now()
	conn := rp.Pool.Get()
	defer conn.Close()
	//conn.Do("SET", key, val)
	conn.Send("SET", key, val)
	conn.Flush()
	conn.Receive()
    log.Println("redisSet consumed：", time.Now().Sub(startTime))
}

func (rp *RedisPool) RedisGet( key string) string {
    startTime := time.Now()
	conn := rp.Pool.Get()
	defer conn.Close()
	conn.Send("GET", key)
	conn.Flush()
	val, err := redis.String(conn.Receive())
	log.Println("redisGet consumed：", time.Now().Sub(startTime))
	if err == nil {
		return val
	}else{
		log.Println(err)
		return ""
	}
}


func ConfReader(path string) map[string]interface{} {
	var conf=make( map[string]interface{} )
	f, _ := os.Open(path)
	buf := bufio.NewReader(f)
	for {
		l, err := buf.ReadString('\n')
		line := strings.TrimSpace(l)
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			if len(line) == 0 {
				break
			}
		}
		switch {
		case len(line) == 0:
			case line[0] == '[' && line[len(line)-1] == ']':
				//session  "[db]"
			section := strings.TrimSpace(line[1 : len(line)-1])
			_ = section
		default:
			i := strings.IndexAny(line, "=")
			conf[strings.TrimSpace(line[0:i])] = strings.TrimSpace(line[i+1:])

		}
	}
	log.Println("Check configuration : ")
	for k, v := range conf {
		log.Println(k," => ", v)
	}

	return conf
}

type Flow struct{
	Mutex *sync.Mutex
	Total *int
	TcpMaxConn *int
}

func ( f *Flow ) Acquire() bool {

	if *f.Total > *f.TcpMaxConn {
		return false
	}

	f.Mutex.Lock()
	*f.Total ++
	f.Mutex.Unlock()
	log.Println("conn number : ", *f.Total)
	return true;
}

func ( f *Flow ) Release() {
	f.Mutex.Lock()
	*f.Total --
	f.Mutex.Unlock()
	log.Println("conn number : ", *f.Total)
}


type MysqlCli struct{
	DBUser *string;
	DBPass *string;
	DBName *string;
	DBAddr *string;
	MDB *sql.DB
}

var mCli *MysqlCli

func (my *MysqlCli ) Connect() bool {
	if  my.MDB == nil{
		var err error
		/*
		dbUser := conf["db_user"].(string)
		dbPass := conf["db_pass"].(string)
		dbName := conf["db_name"].(string)
		dbAddr := conf["mysql_host"].(string) + ":" + conf["mysql_port"].(string)
		*/
		my.MDB, err = sql.Open("mysql", *my.DBUser + ":" + *my.DBPass + "@tcp(" + *my.DBAddr + ")/" + *my.DBName)
		if err != nil {
			log.Println(err.Error())
			return false
		}
	}
	return true
}

func (my *MysqlCli) Close(){
	if my.MDB != nil {
		my.MDB.Close()
	}
}

func (my *MysqlCli) Inquery(sql string, args  ...interface{} ) bool{
	my.Connect()
	stmt, err := my.MDB.Prepare(sql)
	_, err =stmt.Exec(args...)
	if err == nil{
		return true
	}else{
		return false
	}
}

