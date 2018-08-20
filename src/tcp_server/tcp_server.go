package main

import (
	_ "github.com/go-sql-driver/mysql"
	"os"
	"log"
	"strconv"
	"strings"
	"net"
	"net/rpc"
	"util"
	"time"
	"runtime"
	"bufio"
	"sync"
	"encoding/json"
)

var commType = "tcp" //default tcp, can be rpc

var conf = make(map[string]interface{})

var logDir string

var globalLogFile *os.File

var mFlow *util.Flow

var mCli *util.MysqlCli

var redisPool *util.RedisPool

func insertAvatar( uuid string, pid string) string {
	startTime := time.Now()
	log.Println("avatar paras",uuid,pid)
	//sql := "insert into  avatar (uuid,pid)  values (?,?)"
	affect := mCli.Inquery("insert avatar set uuid=?, pid= ?", string(uuid), string(pid))
	log.Println("insertAvatar consumed：", time.Now().Sub(startTime))
	if affect  {
		redisPool.RedisSet("uuid_pid:"+uuid,pid)
		return "{\"code\":0,\"msg\":\"success\",\"data\":\"\"}";
	} else {
		return "{\"code\":2,\"msg\":\"failed to insert avatar\"}";
	}
}

func insertUser( realname string, nickname string, pwd string, avatar string) string {
	startTime := time.Now()
	resp := redisPool.RedisGet( "user:" + realname)
	if resp != "" {
		log.Println(realname +" already exists!")
		return "{\"code\":1,\"msg\":\"should NOT overwrite existing data\",\"uuid\":\"\"}"
	}

	uuid := util.UUID()
	hashedPwd :=string(util.Hash(pwd))
	mCli.Inquery("INSERT user SET uuid=?,realname=?,nickname=?,pwd=?",uuid, realname,nickname,hashedPwd)
	if mCli.MDB == nil{
		log.Println("mysql client is nil")
	}

	redisPool.RedisSet("user:"+realname, uuid + "_"+ hashedPwd + "_" + nickname)
	redisPool.RedisSet("uuid:"+uuid, uuid + "_"+ hashedPwd + "_" + nickname+ "_" + realname)
	log.Println("insertUser consumed：", time.Now().Sub(startTime))
	return login(realname ,pwd)
}

func login(realname string, pwd string) string {
    startTime := time.Now()
	hashedPwd :=string(util.Hash(pwd))
	resp := redisPool.RedisGet("user:"+realname)
	if resp == "" {
		return "{\"code\":1,\"msg\":\"fail\",\"uuid\":\"\"}"
	}
	uuidPwdNickname := strings.Split(resp,"_")
	log.Println("check upn: " + resp)
	log.Println("check uuid: " + uuidPwdNickname[0] )
	log.Println("check pwd: " + uuidPwdNickname[1] )
	log.Println("check nn: " + uuidPwdNickname[2] )

	if hashedPwd != uuidPwdNickname[1]{
		return "{\"code\":1,\"msg\":\"failed\",\"uuid\":\"\"}"
	}
    log.Println("login consumed：", time.Now().Sub(startTime))
	return "{\"code\":0,\"msg\":\"success\",\"uuid\":\"" + uuidPwdNickname[0] + "\"}"
}

func lookup(uuid string) string {
    startTime := time.Now()
	photoID := redisPool.RedisGet("uuid_pid:"+uuid)
	if photoID ==""{
		return "{\"code\":2,\"msg\":\"failed\",\"nickname\":\"\",\"photoid\":\"" + photoID + "\"}"
	}
	resp := redisPool.RedisGet("uuid:"+uuid)
	if resp == "" {
		return "{\"code\":3,\"msg\":\"failed\",\"nickname\":\"\",\"photoid\":\"" + photoID + "\"}"
	}
	idPwdPidNnRn := strings.Split(resp,"_")
    log.Println("lookup consumed：", time.Now().Sub(startTime))
	return "{\"code\":0,\"msg\":\"success\",\"nickname\":\"" + idPwdPidNnRn[2] +"\",\"photoid\":\"" + photoID + "\"}"
}

func lookupAvatar(uuid string) string {
	startTime := time.Now()
	resp := redisPool.RedisGet("uuid_pid:" +uuid)
	log.Println("lookup avatar : " + resp)
	log.Println("lookupAvatar consumed：", time.Now().Sub(startTime))
	return "{\"code\":0,\"msg\":\"success\",\"photoid\":\"" + resp + "\"}"
}

func updateNickname( uuid string, nickname string) string {
	startTime := time.Now()
	mCli.Inquery("update user set nickname=? where uuid=?",nickname, uuid)
	uuidPidNnRn := redisPool.RedisGet("uuid:"+uuid)
	log.Println(uuidPidNnRn)
	upnr := strings.Split(uuidPidNnRn,"_")
	redisPool.RedisSet("uuid:"+uuid, uuid + "_" + upnr[1] + "_" + nickname+ "_" + upnr[3])
	log.Println(upnr[0])
	log.Println(upnr[1])
	log.Println(upnr[2])
	log.Println(upnr[3])

	uuidPwdNn := redisPool.RedisGet("uuid:" + upnr[3])
	upn := strings.Split(uuidPwdNn,"_")
	log.Println("upn: " + uuidPwdNn)
	_ = upn
	log.Println("updateNickname consumed：", time.Now().Sub(startTime))
	return "{\"code\":0,\"msg\":\"\"}";
}

func updateAvatar( uuid string, pid string) string {
	startTime := time.Now()
	affect := mCli.Inquery("update avatar set pid=? where uuid=?",pid, uuid)
	log.Println("updateAvatar consumed：", time.Now().Sub(startTime))
	if affect {
		return "{\"code\":0,\"msg\":\"success\"}";
	} else {
		return "{\"code\":1,\"msg\":\"failed to update avatar\"}";
	}
}

func businessLogics(paras []string) string {
	if paras == nil || len(paras) == 0 {
		log.Println("Parameter error : nil or empty")
		return ""
	}
	var reply string
	switch paras[0] {
		case "SignUp":
			if len(paras) < 5 {
				log.Println("Parameter error")
				return ""
			}
			return insertUser(paras[1], paras[2], paras[3], paras[4])+"\n"
		case "InitAvatar":
			if len(paras) < 3 {
				log.Println("Parameter error")
				return ""
			}
			reply = insertAvatar(paras[1], paras[2])+"\n"
		case "ChangeNickname":
			if len(paras) < 3 {
				log.Println("Parameter error")
				return ""
			}
			reply = updateNickname(paras[1], paras[2])+"\n"
		case "Lookup":
			if len(paras) < 2 {
				log.Println("Parameter error")
				return ""
			}
			reply = lookup(paras[1]) +"\n"
		case "SignIn":
			if len(paras) < 3 {
				log.Println("Parameter error")
				return ""
			}
			reply = login(paras[1], paras[2]) +"\n"
	}
	return reply
}


// Handles incoming requests.
func tcpRequestHandler(conn net.Conn) {
	conn.SetDeadline(time.Now().Add( 5 * time.Second))
	ipStr := conn.RemoteAddr().String()
	log.Println("connected :" + ipStr)
	defer func() {
		log.Println("disconnected :" + ipStr)
		conn.Close()
		mFlow.Release()
	}()
	reader := bufio.NewReader(conn)
	msg, err := util.Decode(reader)
	if err != nil {
		log.Println("Error reading:", err)
		conn.Close()
		mFlow.Release()
		return
	}

	log.Println(conn.RemoteAddr().String() + ":" + string(msg))
	reply := "{\"code\":1,\"msg\":\"para error\",\"uuid\":\"\"}"
	var paras []string
	err = json.Unmarshal([]byte(msg), &paras)
	if err != nil {
		log.Println(err)
		conn.Write([]byte(reply))
		conn.Close()
		mFlow.Release()
		return
	}
	reply = businessLogics(paras)
	log.Println("server resp :", reply)
	resp,err:=util.Encode(reply)
	_,err = conn.Write(resp)
	if err != nil {
		log.Println(err)
		conn.Close()
		mFlow.Release()
		return
	}
}


type query string

func (t *query) SignUp( args *util.Args4, reply *string) error{
	*reply = insertUser(args.A, args.B, args.C, args.D)
	return nil
}

func (t *query) SignIn( args *util.Args2, reply *string) error{
	*reply = login(args.A, args.B)
	return nil
}

func (t *query) Lookup( args *util.Args2, reply *string) error{
	*reply = lookup(args.A)
	return nil
}

func (t *query) LookupAvatar( args *util.Args2, reply *string) error{
	*reply = lookupAvatar(args.A)
	return nil
}

func (t *query) InitAvatar( args *util.Args2, reply *string) error{
	*reply = insertAvatar(args.A, args.B)
	return nil
}

func (t *query) ChangeAvatar( args *util.Args2, reply *string) error{
	*reply = updateAvatar(args.A, args.B)
	return nil
}

func (t *query) ChangeNickname( args *util.Args2, reply *string) error{
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
	dbUser := conf["db_user"].(string)
	dbPass := conf["db_pass"].(string)
	dbName := conf["db_name"].(string)
	dbAddr := conf["mysql_host"].(string) + ":" + conf["mysql_port"].(string)

	mCli = &util.MysqlCli{
		MDB : nil,
		DBUser : &dbUser,
		DBPass : &dbPass,
		DBName : &dbName,
		DBAddr : &dbAddr,
	}
	mCli.Connect()

	redisHost := conf["redis_host"].(string)
	redisPort := conf["redis_port"].(string)
	redisAddr := redisHost + ":" + redisPort
	maxIdle, err := strconv.Atoi(conf["redis_max_idle"].(string))
	maxActive, err := strconv.Atoi(conf["redis_max_active"].(string))
	log.Println("redis addr : " + redisAddr)
	//initRedis("tcp", redisAddr)
	redisPool = &util.RedisPool{
		MaxIdle : &maxIdle,
		MaxActive : &maxActive,
		Addr : &redisAddr,
		Pool : nil,
	}
	redisPool.NewPool()

	commType = conf["proto"].(string)
	log.Println("communication type  : " + commType)
	maxConn, err := strconv.Atoi(conf["tcp_max_conn"].(string))
	log.Println("max tcp conn number  : " , maxConn)
	num := 0

	mFlow = &util.Flow{
		Mutex : &sync.Mutex{},
		Total : &num,
		TcpMaxConn : &maxConn,
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
	tcpPort := conf["tcp_server_port"].(string)

	if commType == "rpc" {
		teller := new(query)
		rpc.Register(teller)
		tcpAddr, err := net.ResolveTCPAddr("tcp", ":" + tcpPort)
		listener, err := net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			log.Println("Error listening:", err.Error())
			os.Exit(1)
		}

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Error accepting: ", err.Error())
				continue
			}
			rpc.ServeConn(conn)
		}
	}

	if commType == "tcp" {
		l, err := net.Listen("tcp", ":" + tcpPort)
		if err != nil {
			log.Println("Error listening:", err.Error())
			os.Exit(1)
		}

		defer l.Close()
		log.Println("listening on ", tcpPort)
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println("Error accepting: ", err.Error())
				continue
			}
			if mFlow.Acquire() {
				go tcpRequestHandler(conn)
			}else{
				//downgrade service ,should do something
				log.Println("Warning : tcp conn too many !")
			}
		}
	}
}


