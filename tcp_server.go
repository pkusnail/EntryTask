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
    )

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

func insertUser( realname string, nickname string, pwd string, avatar string) bool {
	uuid := uuID()
	//db, err := sql.Open("mysql", "root:HappyAlejandroSeaah999@tcp(198.13.43.63:3306)/UserDB?tls=skip-verify&autocommit=true")
	db :=dbConn()
	//checkErr(err)
	// check if realname already exists ,if so , insert fail
        
	var idNum int
        sqlStatement := `SELECT count(*) FROM user  WHERE realname=?`
        row := db.QueryRow(sqlStatement, realname)
        err := row.Scan(&idNum)
        if idNum > 0 {
                return false; //should NOT overwrite existing data
        }

	stmt, err := db.Prepare("INSERT user SET uuid=?,realname=?,nickname=?,pwd=?")
	checkErr(err)
	hashedPwd :=string(hash(pwd)) 
	fmt.Println(hashedPwd)
	es, err := stmt.Exec(uuid, realname,nickname,hashedPwd)
	_ = es
	fmt.Println(err)
	return true;
}

func login(realname string, pwd string) string {
	hashedPwd :=string(hash(pwd)) 
	db := dbConn()

	var uuid string
	sqlStatement := `SELECT uuid FROM user WHERE realname=? and pwd=?`
	row := db.QueryRow(sqlStatement, realname, hashedPwd)
	err := row.Scan(&uuid)
	if err != nil {
	    if err == sql.ErrNoRows {
		//fmt.Println("Zero rows found")
		return ""
	    } else {
		panic(err)
	    }
	}
	return uuid
}

func updateNickname( uuid string, nickname string) string {
	db := dbConn()
	
	var idNum int
        sqlStatement := `SELECT count(*) FROM user WHERE uuid=?`
        row := db.QueryRow(sqlStatement, uuid)
        err := row.Scan(&idNum)
	if idNum < 1 {
		return "{code:1, msg:' user NOT exists'}";
	}
        if err != nil {
            if err == sql.ErrNoRows {
                return "{code:2, msg:'No row found'}";
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
	return "{code:0, msg:''}";
}


func insertAvatar( uuid string, pid string) string {
	db := dbConn()
	//check uuid
	var idNum int
        sqlStatement := `SELECT count(*) FROM avatar  WHERE uuid=?`
        row := db.QueryRow(sqlStatement, uuid)
        err := row.Scan(&idNum)
	if idNum > 0 {
		return "{code:1, msg:'already exists'}"; //should NOT overwrite existing data
	}

	stmt, err := db.Prepare("insert into avatar set  uuid=?,pid=?")
	checkErr(err)
	res, err := stmt.Exec(uuid,pid)
	checkErr(err)
	affect, err := res.RowsAffected()
	if affect > 0 {
		return "{code:0, msg:'success'}";
	} else {
		return "{code:2, msg:'failed to insert avatar'}";
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
		return "{code:0, msg:'success'}";
	} else {
		return "{code:1, msg:'failed to update avatar'}";
	}
	//checkErr(err)
}


func main() {
	/*fmt.Println(insertUser("kljrealabcd","nick","pwd","avatar"))
	fmt.Println(updateNickname( "7c6ff5a0-137f-484a-bc71-ab63d7b1d9b4", "ppptring"))
	fmt.Println(login("kljrcd","pwd"))
	fmt.Println(insertAvatar( "abcd", "pid string"))
	fmt.Println(updateAvatar( "abcd", "pitring"))
	*/
		
}
