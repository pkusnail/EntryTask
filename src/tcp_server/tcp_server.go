package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"os"
	"log"
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
	"sync"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

var commType = "tcp" //default tcp, can be rpc

var conf = make(map[string]interface{})

var logDir string

var globalLogFile *os.File

var tcpMaxConn = 10000


type flow struct{
	mutex *sync.Mutex
	total *int
}

func ( f *flow ) acquire() bool {
	if f.mutex == nil{
		f.mutex = &sync.Mutex{}
		num := 0
		f.total = &num
	}

	if *f.total > tcpMaxConn {
		return false
	}

	f.mutex.Lock()
	*f.total ++
	f.mutex.Unlock()
	log.Println("conn number : ", *f.total)
	return true;
}

func ( f *flow ) release() {
	f.mutex.Lock()
	*f.total --
	f.mutex.Unlock()
	log.Println("conn number : ", *f.total)
}

var mFlow *flow


var cq = make(chan net.Conn, tcpMaxConn) // http client conn queue, default 10000

type mysqlCli struct{
	db *sql.DB
}

var mCli *mysqlCli

func (my *mysqlCli ) Connect() bool {
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
			return false
		}
	}
	return true
}

func (my *mysqlCli) Close(){
	if my.db != nil {
		my.db.Close()
	}
}

//func (my *mysqlCli) Inquery(sql string, paras  ... string ) bool{
func (my *mysqlCli) Inquery(sql string, paras  ...interface{} ) bool{
	my.Connect()
	stmt, err := my.db.Prepare(sql)
	_, err =stmt.Exec(paras...)
	if err == nil{
		return true
	}else{
		return false
	}
}


var redisMaxConn = 20
var redisAddr = "127.0.0.1:6379" //default value
var redisPoll chan redis.Conn

func putRedis(conn redis.Conn) {
    if redisPoll == nil {
        redisPoll = make(chan redis.Conn, redisMaxConn)
    }
    if len(redisPoll) >= redisMaxConn {
        conn.Close()
        return
    }
    redisPoll <- conn
}

func initRedis(network, address string) redis.Conn {
    if len(redisPoll) == 0 {
        redisPoll = make(chan redis.Conn, redisMaxConn)
        go func() {
            for i := 0; i < redisMaxConn; i++ {
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
    c := initRedis("tcp", redisAddr)
	c.Do("SET", key, val)
    log.Println("redisSet consumed：", time.Now().Sub(startTime))
}

func redisGet(key string) string  {
    startTime := time.Now()
    c := initRedis("tcp", redisAddr)
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


func insertAvatar( uuid string, pid string) string {
	startTime := time.Now()
	log.Println("avatar paras",uuid,pid)
	//sql := "insert into  avatar (uuid,pid)  values (?,?)"
	affect := mCli.Inquery("insert avatar set uuid=?, pid= ?", string(uuid), string(pid))
	log.Println("insertAvatar consumed：", time.Now().Sub(startTime))
	if affect  {
		redisSet("uuid_pid:"+uuid,pid)
		return "{\"code\":0,\"msg\":\"success\",\"data\":\"\"}";
	} else {
		return "{\"code\":2,\"msg\":\"failed to insert avatar\"}";
	}
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
	// lookup the redis cache first
	photoID := redisGet("uuid_pid:"+uuid)
	if photoID ==""{
		return "{\"code\":2,\"msg\":\"failed\",\"nickname\":\"\",\"photoid\":\"" + photoID + "\"}"
	}
	resp := redisGet("uuid:"+uuid)
	if resp == "" {
		return "{\"code\":3,\"msg\":\"failed\",\"nickname\":\"\",\"photoid\":\"" + photoID + "\"}"
	}
	idPwdPidNnRn := strings.Split(resp,"_")
    log.Println("lookup consumed：", time.Now().Sub(startTime))
	return "{\"code\":0,\"msg\":\"success\",\"nickname\":\"" + idPwdPidNnRn[2] +"\",\"photoid\":\"" + photoID + "\"}"
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
	uuidPidNnRn := redisGet("uuid:"+uuid)
	log.Println(uuidPidNnRn)
	upnr := strings.Split(uuidPidNnRn,"_")
	redisSet("uuid:"+uuid, uuid + "_" + upnr[1] + "_" + nickname+ "_" + upnr[3])
	log.Println(upnr[0])
	log.Println(upnr[1])
	log.Println(upnr[2])
	log.Println(upnr[3])

	uuidPwdNn := redisGet("uuid:" + upnr[3])
	upn := strings.Split(uuidPwdNn,"_")
	log.Println("upn: " + uuidPwdNn)
	_=upn
	log.Println("updateNickname consumed：", time.Now().Sub(startTime))
	return "{\"code\":0,\"msg\":\"\"}";
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
		mFlow.release()
	}()
	reader := bufio.NewReader(conn)
	msg, err := util.Decode(reader)
	if err != nil {
		log.Println("Error reading:", err)
		conn.Close()
		mFlow.release()
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
		mFlow.release()
		return
	}
	reply = businessLogics(paras)
	log.Println("server resp :", reply)
	resp,err:=util.Encode(reply)
	_,err = conn.Write(resp)
	if err != nil {
		log.Println(err)
		conn.Close()
		mFlow.release()
		return
	}
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

	redisMaxConn, err = strconv.Atoi(conf["redis_max_conn"].(string))
	if err != nil{
		log.Println(err)
	}
	redisHost := conf["redis_host"].(string)
	redisPort := conf["redis_port"].(string)
	redisAddr = redisHost + ":" + redisPort
	log.Println("redis addr : " + redisAddr)
	commType = conf["proto"].(string)
	log.Println("communication type  : " + commType)
	tcpMaxConn, err = strconv.Atoi(conf["tcp_max_conn"].(string))
	log.Println("max tcp conn number  : " , tcpMaxConn)
	mFlow = &flow{mutex:nil,total:nil}
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
	tcpPort := conf["tcp_server_port"].(string)

	if commType == "rpc" {
		teller := new(query)
		rpc.Register(teller)
		tcpAddr, err := net.ResolveTCPAddr("tcp", ":" + tcpPort)
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
			if mFlow.acquire() {
				go tcpRequestHandler(conn)
			}else{
				//downgrade service ,should do something
				log.Println("Warning : tcp conn too many !")
			}
		}
	}
}


