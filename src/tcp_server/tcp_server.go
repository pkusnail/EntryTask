package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"os"
	"log"
	"fmt"
	"os/exec"
	"hash/fnv"
	"strconv"
	"strings"
	"net"
	"net/rpc"
	"util"
	"time"
	"runtime"
	"bufio"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

var commType = "tcp" //default tcp, can be rpc

var conf = make(map[string]interface{})

var logDir string

var globalLogFile *os.File

var TCP_MAX_CONN = 10000

var cq = make(chan net.Conn, TCP_MAX_CONN) // http client conn queue, default 10000

type mysqlCli struct{
	db *sql.DB
}

var mCli *mysqlCli

func (my *mysqlCli ) Connect() {
	if  my.db == nil{
		var err error
		dbDriver := "mysql"
		dbUser := conf["db_user"].(string)
		dbPass := conf["db_pass"].(string)
		dbName := conf["db_name"].(string)
		dbAddr := conf["mysql_host"].(string) + ":" + conf["mysql_port"].(string) 
		my.db, err = sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(" + dbAddr +")/"+dbName)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func (my *mysqlCli) Close(){
	if my.db != nil {
		my.db.Close()
	}
}

func (my *mysqlCli) Inquery(sql string, paras ... string ) bool{
	my.Connect()
	stmt, err := my.db.Prepare(sql)
	if len(paras) == 1 {
		_, err =stmt.Exec(paras[0])
	}else if len(paras) == 2 {
		_, err =stmt.Exec(paras[0], paras[1])
	}else if len(paras) == 3 {
		_, err =stmt.Exec(paras[0],paras[1] , paras[2])
	}else if len(paras) == 4 {
		_, err =stmt.Exec(paras[0],paras[1] , paras[2], paras[3])
	}

	if err == nil{
		return true
	}else{
		return false
	}
}


var REDIS_MAX_CONN = 20
var REDIS_ADDR = "127.0.0.1:6379" //default value
var redisPoll chan redis.Conn

func putRedis(conn redis.Conn) {
    if redisPoll == nil {
        redisPoll = make(chan redis.Conn, REDIS_MAX_CONN)
    }
    if len(redisPoll) >= REDIS_MAX_CONN {
        conn.Close()
        return
    }
    redisPoll <- conn
}

func initRedis(network, address string) redis.Conn {
    if len(redisPoll) == 0 {
        redisPoll = make(chan redis.Conn, REDIS_MAX_CONN)
        go func() {
            for i := 0; i < REDIS_MAX_CONN; i++ {
                c, err := redis.Dial(network, address)
                if err != nil {
                    panic(err)
                }
                putRedis(c)
            }
        } ()
    }
    return <-redisPoll
}

func redisSet(key string, val string)  {
    startTime := time.Now()
    c := initRedis("tcp", REDIS_ADDR)
	c.Do("SET", key, val)
    log.Println("redisSet consumed：", time.Now().Sub(startTime))
}

func redisGet(key string) string  {
    startTime := time.Now()
    c := initRedis("tcp", REDIS_ADDR)
	val, _ := redis.String(c.Do("GET", key))
	log.Println("redisGet consumed: ", time.Now().Sub(startTime))
	return val
}

func uuID() string {
    out, err := exec.Command("uuidgen").Output()
    if err != nil {
        log.Fatal(err)
    }
    return strings.Replace(string(out), "\n", "", -1)
}

func hash(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return strconv.FormatUint(h.Sum64(), 10)
}

func insertUser( realname string, nickname string, pwd string, avatar string) string {
	startTime := time.Now()
	//redis format :  username:realname
	resp := redisGet( "user:" + realname)
	if resp != "" {
		log.Println(realname +" already exists!")
		return "{\"code\":1,\"msg\":\"should NOT overwrite existing data\",\"uuid\":\"\"}"
	}

	uuid := uuID()
	hashedPwd :=string(hash(pwd))
	mCli.Inquery("INSERT user SET uuid=?,realname=?,nickname=?,pwd=?",uuid, realname,nickname,hashedPwd)
	if mCli.db == nil{
		log.Println("mysql client is nil")
	}

	redisSet("user:"+realname, uuid + "_"+ hashedPwd + "_" + nickname)
	redisSet("uuid:"+uuid, uuid + "_"+ hashedPwd + "_" + nickname+ "_" + realname)
	log.Println("insertUser consumed：", time.Now().Sub(startTime))
	return login(realname ,pwd)
}

func login(realname string, pwd string) string {
    startTime := time.Now()
	hashedPwd :=string(hash(pwd))
	resp := redisGet("user:"+realname)
	if resp == "" {
		return "{\"code\":1,\"msg\":\"fail\",\"uuid\":\"\"}"
	}
	uuid_pwd_nickname := strings.Split(resp,"_")
	log.Println("check upn: " + resp)
	log.Println("check uuid: " + uuid_pwd_nickname[0] )
	log.Println("check pwd: " + uuid_pwd_nickname[1] )
	log.Println("check nn: " + uuid_pwd_nickname[2] )

	if hashedPwd != uuid_pwd_nickname[1]{
		return "{\"code\":1,\"msg\":\"failed\",\"uuid\":\"\"}"
	}
    log.Println("login consumed：", time.Now().Sub(startTime))
	return "{\"code\":0,\"msg\":\"success\",\"uuid\":\"" + uuid_pwd_nickname[0] + "\"}"
}

func lookup(uuid string) string {
    startTime := time.Now()
	// lookup the redis cache first
	photoID := redisGet("uuid_pid:"+uuid)
	if photoID ==""{
		return "{\"code\":2,\"msg\":\"failed\",\"nickname\":\"\",\"photoid\":\"" + photoID + "\"}"
	}
	resp := redisGet("uuid:"+uuid)
	if resp == "" {
		return "{\"code\":3,\"msg\":\"failed\",\"nickname\":\"\",\"photoid\":\"" + photoID + "\"}"
	}
	id_pwd_pid_nn_rn := strings.Split(resp,"_")
    log.Println("lookup consumed：", time.Now().Sub(startTime))
	return "{\"code\":0,\"msg\":\"success\",\"nickname\":\"" + id_pwd_pid_nn_rn[2] +"\",\"photoid\":\"" + photoID + "\"}"
}

func lookupAvatar(uuid string) string {
	startTime := time.Now()
	resp := redisGet("uuid_pid:" +uuid)
	log.Println("lookup avatar : " + resp)
	log.Println("lookupAvatar consumed：", time.Now().Sub(startTime))
	//return "{code:0,msg :'success',data:'{uuid:" + uuid + "}'}"
	return "{\"code\":0,\"msg\":\"success\",\"photoid\":\"" + resp + "\"}"
}

func updateNickname( uuid string, nickname string) string {
	startTime := time.Now()
	mCli.Inquery("update user set nickname=? where uuid=?",nickname, uuid)
	uuid_pid_nn_rn := redisGet("uuid:"+uuid)
	log.Println(uuid_pid_nn_rn)
	upnr := strings.Split(uuid_pid_nn_rn,"_")
	redisSet("uuid:"+uuid, uuid + "_" + upnr[1] + "_" + nickname+ "_" + upnr[3])
	log.Println(upnr[0])
	log.Println(upnr[1])
	log.Println(upnr[2])
	log.Println(upnr[3])

	uuid_pwd_nn := redisGet("uuid:" + upnr[3])
	upn := strings.Split(uuid_pwd_nn,"_")
	log.Println("upn: " + uuid_pwd_nn)
	_=upn
	log.Println("updateNickname consumed：", time.Now().Sub(startTime))
	return "{\"code\":0,\"msg\":\"\"}";
}

func insertAvatar( uuid string, pid string) string {
	startTime := time.Now()
	log.Println("avatar paras",uuid,pid)
	sql := "insert into  avatar (uuid,pid)  values (?,?)"
	affect := mCli.Inquery(sql, uuid, pid)
	log.Println("insertAvatar consumed：", time.Now().Sub(startTime))
	if affect  {
		redisSet("uuid_pid:"+uuid,pid)
		return "{\"code\":0,\"msg\":\"success\",\"data\":\"\"}";
	} else {
		return "{\"code\":2,\"msg\":\"failed to insert avatar\"}";
	}
}

func updateAvatar( uuid string, pid string) string {
	startTime := time.Now()
	affect := mCli.Inquery("update avatar set pid=? where uuid=?",pid, uuid)
	log.Println("updateAvatar consumed：", time.Now().Sub(startTime))
	if affect {
		//update redis cache	
		return "{\"code\":0,\"msg\":\"success\"}";
	} else {
		return "{\"code\":1,\"msg\":\"failed to update avatar\"}";
	}
}

type Query string

func (t *Query) SignUp( args *util.Args4, reply *string) error{
	*reply = insertUser(args.A, args.B, args.C, args.D)
	return nil
}

func (t *Query) SignIn( args *util.Args2, reply *string) error{
	*reply = login(args.A, args.B)
	return nil
}

func (t *Query) Lookup( args *util.Args2, reply *string) error{
	*reply = lookup(args.A)
	return nil
}

func (t *Query) LookupAvatar( args *util.Args2, reply *string) error{
	*reply = lookupAvatar(args.A)
	return nil
}

func (t *Query) InitAvatar( args *util.Args2, reply *string) error{
	*reply = insertAvatar(args.A, args.B)
	return nil
}

func (t *Query) ChangeAvatar( args *util.Args2, reply *string) error{
	*reply = updateAvatar(args.A, args.B)
	return nil
}

func (t *Query) ChangeNickname( args *util.Args2, reply *string) error{
	*reply = updateNickname(args.A, args.B)
	return nil
}




func init(){
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	conf = util.ConfReader(dir + "/../../conf/setting.conf")
	logDir = conf["log_file_dir"].(string)
	globalLogFile, err = os.OpenFile( dir + "/" + logDir + "/tcp_server.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: ", err)
	}
	defer globalLogFile.Close()
	log.SetOutput(globalLogFile)
	mCli = &mysqlCli{db:nil}

	REDIS_MAX_CONN, err = strconv.Atoi(conf["redis_max_conn"].(string))
	if err != nil{
		log.Println(err)
	}
	REDIS_HOST := conf["redis_host"].(string)
	REDIS_PORT := conf["redis_port"].(string)
	REDIS_ADDR = REDIS_HOST + ":" + REDIS_PORT
	log.Println("redis addr : " + REDIS_ADDR)
	commType = conf["proto"].(string)
	log.Println("communication type  : " + commType)
	
	TCP_MAX_CONN, err = strconv.Atoi(conf["tcp_max_conn"].(string))
	cq = make(chan net.Conn, TCP_MAX_CONN)
	log.Println("max tcp conn number  : " , TCP_MAX_CONN)

}



// Handles incoming requests.
func tcpRequestHandler() {
	for conn := range cq {
		ipStr := conn.RemoteAddr().String()
		defer func() {
			fmt.Println("disconnected :" + ipStr)
			conn.Close()
		}()
		reader := bufio.NewReader(conn)
		msg, err := util.Decode(reader)

		if err != nil {
			fmt.Println("Error reading:", err)
			//conn.Close()
			return
		}
		log.Println(conn.RemoteAddr().String() + ":" + string(msg))
		reply := "{\"code\":1,\"msg\":\"para error\",\"uuid\":\"\"}"
		var paras []string
		err = json.Unmarshal([]byte(msg), &paras)
		if err != nil {
			log.Println(err)
			conn.Write([]byte(reply))
			//conn.Close()
		}

		switch paras[0] {
			case "SignUp":
				reply = insertUser(paras[1], paras[2], paras[3], paras[4])+"\n"
			case "InitAvatar":
				reply = insertAvatar(paras[1], paras[2])+"\n"
			case "ChangeNickname":
				reply = updateNickname(paras[1], paras[2])+"\n"
			case "Lookup":
				reply = lookup(paras[1]) +"\n"
			case "SignIn":
				reply = login(paras[1], paras[2]) +"\n"
		}
		log.Println("server resp :", reply)
		resp,err:=util.Encode(reply)
		conn.Write(resp)
		//conn.Close()
	}
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	globalLogFile, err = os.OpenFile( dir + "/" + logDir + "/tcp_server.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: ", err)
	}
	defer globalLogFile.Close()
	log.SetOutput(globalLogFile)

	runtime.GOMAXPROCS(runtime.NumCPU())
	//tcp_host := conf["tcp_server_host"].(string)
	tcp_port := conf["tcp_server_port"].(string)

	if commType == "rpc" {
		teller := new(Query)
		rpc.Register(teller)
		tcpAddr, err := net.ResolveTCPAddr("tcp", ":" + tcp_port)
		listener, err := net.ListenTCP("tcp", tcpAddr)
		_ = err
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			rpc.ServeConn(conn)
		}
	}

	if commType == "tcp" {
		l, err := net.Listen("tcp", ":" + tcp_port)
		if err != nil {
			log.Println("Error listening:", err.Error())
			os.Exit(1)
		}
		defer l.Close()
		log.Println("listening on ", tcp_port)
		//go tcpRequestHandler()
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println("Error accepting: ", err.Error())
				os.Exit(1)
			}
			if len(cq) < TCP_MAX_CONN {
				cq <- conn
			}else{
				log.Println("Warning : tcp connection queue full !")
				//should do something
			}
			// Handle connections in a new goroutine.
			//go tcpRequestHandler(conn)
			go tcpRequestHandler()
		}
	}
}


