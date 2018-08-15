package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"fmt"
	"os"
	"log"
	"os/exec"
	"hash/fnv"
	"strconv"
	"strings"
	"net"
	"net/rpc"
	"util"
    )

type mysqlCli struct{
	db *sql.DB
}

var mCli *mysqlCli = nil

func Connect()( db *sql.DB, err error){
	if  mCli == nil{
		mCli = new(mysqlCli)
		var err error
		dbDriver := "mysql"
		dbUser := "root"
		dbPass := "HappyAlejandroSeaah999"
		dbName := "UserDB"
		dbAddr := "198.13.43.63:3306"
		mCli.db, err = sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(" + dbAddr +")/"+dbName)
		if err != nil {
			panic(err.Error())
			return nil, err
		}
	}
		return mCli.db ,nil
}

func Close(){
	if mCli != nil {
		mCli.db.Close()
	}
}

func uuID() string {
    out, err := exec.Command("uuidgen").Output()
    if err != nil {
        log.Fatal(err)
    }
    return strings.Replace(string(out), "\n", "", -1)
}

func checkErr(err error) {
    if err != nil {
        fmt.Println("Fatal error ", err.Error())
        os.Exit(1)
    }
}

func hash(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return strconv.FormatUint(h.Sum64(), 10)
}

func insertUser( realname string, nickname string, pwd string, avatar string) string {
	//redis format :  username:realname
	resp := util.RedisGet( "user:" + realname)
	fmt.Println(resp)
	if resp != "" {
		log.Println(realname +" already exists!")
		return "{\"code\":1,\"msg\":\"should NOT overwrite existing data\",\"uuid\":\"\"}"
	}

	uuid := uuID()
	db, err :=Connect()
	stmt, err := db.Prepare("INSERT user SET uuid=?,realname=?,nickname=?,pwd=?")
	checkErr(err)
	hashedPwd :=string(hash(pwd))
	es, err := stmt.Exec(uuid, realname,nickname,hashedPwd)
	_ = es
	log.Println(err)
	util.RedisPut("user:"+realname, uuid + "_"+ hashedPwd + "_" + nickname)
	util.RedisPut("uuid:"+uuid, uuid + "_"+ hashedPwd + "_" + nickname+ "_" + realname)
	return login(realname ,pwd)
}

func login(realname string, pwd string) string {
	hashedPwd :=string(hash(pwd))
	resp := util.RedisGet("user:"+realname)
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
	return "{\"code\":0,\"msg\":\"success\",\"uuid\":\"" + uuid_pwd_nickname[0] + "\"}"
}


func lookup(uuid string) string {
	// lookup the redis cache first
	photoID := util.RedisGet("uuid_pid:"+uuid)
	if photoID ==""{
		return "{\"code\":2,\"msg\":\"failed\",\"nickname\":\"\",\"photoid\":\"" + photoID + "\"}"
	}
	resp := util.RedisGet("uuid:"+uuid)
	if resp == "" {
		return "{\"code\":3,\"msg\":\"failed\",\"nickname\":\"\",\"photoid\":\"" + photoID + "\"}"
	}
	id_pwd_pid_nn_rn := strings.Split(resp,"_")
	return "{\"code\":0,\"msg\":\"success\",\"nickname\":\"" + id_pwd_pid_nn_rn[2] +"\",\"photoid\":\"" + photoID + "\"}"
}


func lookupAvatar(uuid string) string {
    resp := util.RedisGet("uuid_pid:" +uuid)
	log.Println("lookup avatar : " + resp)
	//return "{code:0,msg :'success',data:'{uuid:" + uuid + "}'}"
	return "{\"code\":0,\"msg\":\"success\",\"photoid\":\"" + resp + "\"}"
}

func updateNickname( uuid string, nickname string) string {
	db, _ := Connect()
	var idNum int
	sqlStatement := `SELECT count(*) FROM user WHERE uuid=?`
	row := db.QueryRow(sqlStatement, uuid)
	err := row.Scan(&idNum)
	if idNum < 1 {
		return "{\"code\":1,\"msg\":\"user NOT exists\",\"uuid\":\"\"}";
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return "{\"code\":2,\"msg\":\"No row found\",\"uuid\":\"\"}";
		} else {
			panic(err)
		}
	}
	stmt, err := db.Prepare("update user set nickname=? where uuid=?")
	checkErr(err)
	res, err := stmt.Exec(nickname, uuid)
	checkErr(err)
	affect, err := res.RowsAffected()
	_ = affect
	checkErr(err)
	// update redis cache
	return "{\"code\":0,\"msg\":\"\"}";
}

func insertAvatar( uuid string, pid string) string {
	db,_ := Connect()
	log.Println("inser avatar : " + uuid + " with pid : " + pid)
	stmt, err := db.Prepare("insert into  avatar (uuid,pid)  values (?,?)")
	checkErr(err)
	res, err := stmt.Exec(uuid,pid)
	checkErr(err)
	affect, err := res.RowsAffected()
	if affect > 0 {
		// update redis cache
		util.RedisPut("uuid_pid:"+uuid,pid)
		return "{\"code\":0,\"msg\":\"success\",\"data\":\"\"}";
	} else {
		return "{\"code\":2,\"msg\":\"failed to insert avatar\"}";
	}
}

func updateAvatar( uuid string, pid string) string {
	db, _ := Connect()
	stmt, err := db.Prepare("update avatar set pid=? where uuid=?")
	checkErr(err)
	res, err := stmt.Exec(pid, uuid)
	checkErr(err)
	affect, err := res.RowsAffected()
	_ = affect
	if affect > 0 {
		//update redis cache	
		return "{\"code\":0,\"msg\":\"success\"}";
	} else {
		return "{\"code\":1,\"msg\":\"failed to update avatar\"}";
	}
	//checkErr(err)
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

func main() {

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile( dir + "/../log/tcp_server.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)


    teller := new(Query)
    rpc.Register(teller)

    tcpAddr, err := net.ResolveTCPAddr("tcp", ":9999")
    checkErr(err)

    listener, err := net.ListenTCP("tcp", tcpAddr)
    checkErr(err)

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        rpc.ServeConn(conn)
	}	
}
