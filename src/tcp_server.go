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
	"redis"
    )

func redisPut( key string, val string) bool {
	spec := redis.DefaultSpec().Db(0).Password("")
	client, e := redis.NewSynchClientWithSpec(spec)
	if e != nil {
		log.Println("failed to create redis client", e)
		return false
	}
	value := []byte(val)
	e = client.Set(key, value)
	if e == nil{
		return true
	}
	return false
}

func redisGet( key string) string {
	spec := redis.DefaultSpec().Db(0).Password("")
	client, e := redis.NewSynchClientWithSpec(spec)
	if e != nil {
		log.Println("failed to create the client", e)
		return "NULL"
	}
	value, e := client.Get(key)
	if e != nil {
		log.Println("error on Get", e)
		return "NULL"
	}
	fmt.Println("redisGet: " + string(value[:]))
	return string(value[:])
}

func uuID() string {

    out, err := exec.Command("uuidgen").Output()
    if err != nil {
        log.Fatal(err)
    }
    //fmt.Printf("%s", out)
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
	//return strconv.Itoa(h.Sum64())
}

func dbConn() (db *sql.DB) {
    dbDriver := "mysql"
    dbUser := "root"
    dbPass := "HappyAlejandroSeaah999"
    dbName := "UserDB"
    dbAddr := "198.13.43.63:3306"
    db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(" + dbAddr +")/"+dbName)
    if err != nil {
        panic(err.Error())
    }
    return db
}

func insertUser( realname string, nickname string, pwd string, avatar string) string {
	//redis format :  username:realname
	resp := redisGet( "user:" + realname)
	fmt.Println(resp)
	if resp != "" {
		fmt.Println("Already exists!")
		return "{\"code\":1,\"msg\":\"should NOT overwrite existing data\",\"uuid\":\"\"}"
	}

	uuid := uuID()
	db :=dbConn()
	//checkErr(err)
	// check if realname already exists ,if so , insert fail
	//todo : should replace this part to redis
/*
	var idNum int
	sqlStatement := `SELECT count(*) FROM user  WHERE realname=?`
	row := db.QueryRow(sqlStatement, realname)
	err := row.Scan(&idNum)
	if idNum > 0 {
			return "{\"code\":1,\"msg\":\"should NOT overwrite existing data\"}"
	}
*/
	stmt, err := db.Prepare("INSERT user SET uuid=?,realname=?,nickname=?,pwd=?")
	checkErr(err)
	hashedPwd :=string(hash(pwd))
	fmt.Println(hashedPwd)
	es, err := stmt.Exec(uuid, realname,nickname,hashedPwd)
	_ = es
	//fmt.Println(err)
	redisPut("user:"+realname, uuid + "_"+ hashedPwd + "_" + nickname)
	redisPut("uuid:"+uuid, uuid + "_"+ hashedPwd + "_" + nickname+ "_" + realname)
	return login(realname ,pwd)
	//return "{\"code\":0,\"msg\":\"success\",\"uuid\":\"" + uuid +"\"}";
}

func login(realname string, pwd string) string {
	hashedPwd :=string(hash(pwd))
	//var uuid string
	/*
	db := dbConn()
	var uuid string
	sqlStatement := `SELECT uuid FROM user WHERE realname=? and pwd=?`
	row := db.QueryRow(sqlStatement, realname, hashedPwd)
	err := row.Scan(&uuid)
	if err != nil {
	    if err == sql.ErrNoRows {
		//fmt.Println("Zero rows found")
			return "{\"code\":1,\"msg\":\"failed\",\"data\":\"\"}"
	    } else {
			panic(err)
	    }
	}
	*/
	resp := redisGet("user:"+realname)
	if resp == "" {
		return "{\"code\":1,\"msg\":\"fail\",\"uuid\":\"\"}"
	}
	uuid_pwd_nickname := strings.Split(resp,"_")
	fmt.Println("check upn: " + resp)

	if hashedPwd != uuid_pwd_nickname[1]{
		return "{\"code\":1,\"msg\":\"failed\",\"uuid\":\"\"}"
	}
	//return "{code:0,msg :'success',data:'{uuid:" + uuid + "}'}"
	return "{\"code\":0,\"msg\":\"success\",\"uuid\":\"" + uuid_pwd_nickname[2] + "\"}"
}


func lookup(uuid string) string {
	// lookup the redis cache first
	/*
	// if Not found , just return, deprected
	db := dbConn()
	var nname string
	var photoID string
	sqlStatement := `SELECT nickname FROM user WHERE uuid=?`
	row := db.QueryRow(sqlStatement, uuid)
	err := row.Scan(&nname)
	if err != nil {
	    if err == sql.ErrNoRows {
		//fmt.Println("Zero rows found")
			return "{\"code\":1,\"msg\":\"failed\",\"data\":\"\"}"
	    } else {
			panic(err)
	    }
	}	
	sqlStatement = `SELECT pid FROM avatar WHERE uuid=?`
	row = db.QueryRow(sqlStatement, uuid)
	err = row.Scan(&photoID)
	if err != nil {
	    if err == sql.ErrNoRows {
		//fmt.Println("Zero rows found")
			return "{\"code\":2,\"msg\":\"photo id failed\",\"data\":\"\"}"
	    } else {
			panic(err)
	    }
	}
	*/
	photoID := redisGet("uuid_pid:"+uuid)
	if photoID ==""{
		return "{\"code\":2,\"msg\":\"failed\",\"nickname\":\"\",\"photoid\":\"" + photoID + "\"}"
	}
	fmt.Println("see what: " + uuid)
	resp := redisGet("uuid:"+uuid)
	if resp == "" {
		return "{\"code\":3,\"msg\":\"failed\",\"nickname\":\"\",\"photoid\":\"" + photoID + "\"}"
	}

	id_pwd_pid_rn := strings.Split(resp,"_")
	return "{\"code\":0,\"msg\":\"success\",\"nickname\":\"" + id_pwd_pid_rn[3] +"\",\"photoid\":\"" + photoID + "\"}"
}


func lookupAvatar(uuid string) string {
	
	
	// find it in redis cache

	// if NOT found

	db := dbConn()
	var photoID string
	sqlStatement := `SELECT pid FROM avatar WHERE uuid=?`
	row := db.QueryRow(sqlStatement, uuid)
	err := row.Scan(&photoID)
	if err != nil {
	    if err == sql.ErrNoRows {
		//fmt.Println("Zero rows found")
			return "{\"code\":1,\"msg\":\"photo id failed\",\"data\":\"\"}"
	    } else {
			panic(err)
	    }
	}
	//return "{code:0,msg :'success',data:'{uuid:" + uuid + "}'}"
	return "{\"code\":0,\"msg\":\"success\",\"photoid\":\"" + photoID + "\"}"
}


func updateNickname( uuid string, nickname string) string {
	db := dbConn()
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
	db := dbConn()
	//check uuid
/*	var idNum int
        sqlStatement := `SELECT count(*) FROM avatar  WHERE uuid=?`
        row := db.QueryRow(sqlStatement, uuid)
        err := row.Scan(&idNum)
	if idNum > 0 {
		return "{\"code\":1,\"msg\":\"already exists\"}"; //should NOT overwrite existing data
	}
*/
	stmt, err := db.Prepare("insert into  avatar (uuid,pid)  values (?,?)")
	checkErr(err)
	res, err := stmt.Exec(uuid,pid)
	checkErr(err)
	affect, err := res.RowsAffected()
	if affect > 0 {
		// update redis cache
		redisPut("uuid_pid:"+uuid,pid)
		return "{\"code\":0,\"msg\":\"success\",\"data\":\"\"}";
	} else {
		return "{\"code\":2,\"msg\":\"failed to insert avatar\"}";
	}
	//checkErr(err)
}


func updateAvatar( uuid string, pid string) string {
	db := dbConn()
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

//===================
type Args2 struct {
	A,B string
}

type Args3 struct {
	A,B,C string
}


type Args4 struct {
	A,B,C,D string
}


type Query string

func (t *Query) SignUp( args *Args4, reply *string) error{
	*reply = insertUser(args.A, args.B, args.C, args.D)
	return nil
}

func (t *Query) SignIn( args *Args2, reply *string) error{
	*reply = login(args.A, args.B)
	return nil
}



func (t *Query) Lookup( args *Args2, reply *string) error{
	*reply = lookup(args.A)
	return nil
}


func (t *Query) LookupAvatar( args *Args2, reply *string) error{
	*reply = lookupAvatar(args.A)
	return nil
}
func (t *Query) InitAvatar( args *Args2, reply *string) error{
	*reply = insertAvatar(args.A, args.B)
	return nil
}

func (t *Query) ChangeAvatar( args *Args2, reply *string) error{
	*reply = updateAvatar(args.A, args.B)
	return nil
}

func main() {
	/*fmt.Println(insertUser("kljrealabcd","nick","pwd","avatar"))
	fmt.Println(updateNickname( "7c6ff5a0-137f-484a-bc71-ab63d7b1d9b4", "ppptring"))
	fmt.Println(login("kljrcd","pwd"))
	fmt.Println(insertAvatar( "abcd", "pid string"))
	fmt.Println(updateAvatar( "abcd", "pitring"))
	*/
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
