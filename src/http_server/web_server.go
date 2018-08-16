package main

import (
	"fmt"
	"html/template"
	"log"
	"net/rpc"
	"strings"
	"time"
	"strconv"
	"bytes"
	"io"
//	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"crypto/md5"
	"encoding/json"
	"github.com/gorilla/sessions"
	"util"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "9999"
	CONN_TYPE = "tcp"
)

type Register struct {
	realname string
	nickname string
	info string
}

type HomeInfo struct {
	AvatarUrl string
	Nickname string
}

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)


func  getMillSec() int64{ // return timestamp
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var rname, nname, info  []string
		var ok bool
		pnum := len(r.URL.Query())
		if pnum > 0 {
			info, ok = r.URL.Query()["info"]
		}
		if pnum > 1 {
			rname, ok = r.URL.Query()["rname"]
		}
		if pnum > 2{
			nname, ok = r.URL.Query()["nname"]
		}

		popup := strings.Join(info,"")
		rn := strings.Join(rname,"")
		nn := strings.Join(nname,"")
		_ = ok
		data := Register {
			realname : rn,
			nickname : nn,
			info : popup,
		}
		t, _ := template.ParseFiles("tpl/signup.gtpl")
		t.Execute(w, data)
	} else {
		r.ParseForm()
		// logic part of signup
		realname := strings.Join(r.Form["rname"],"")
		nickname := strings.Join(r.Form["nname"],"")
		pwd := strings.Join(r.Form["pwd"],"")
		//communicate with tcp server and proxy server  
		client, err := rpc.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
		if err != nil {
			log.Println(err)
			log.Fatal("dialing:", err)
		}
		args := util.Args4{realname,nickname,pwd,""}
		var reply string
		err = client.Call("Query.SignUp", args, &reply)
		client.Close()
		if err != nil {
			log.Println(err)
			log.Fatal("Query.SignUp error:", err)
		}
		log.Println("check : %v",reply)	
		byt := []byte(reply)
		var dat map[string]interface{}
		if err := json.Unmarshal(byt, &dat); err != nil {
			log.Println("careful :" + reply)
			panic(err)
		}
		code := dat["code"].(float64)
		uuid := dat["uuid"].(string)
		msg := dat["msg"].(string)
		if code != 0 {
			log.Println("Sign up failed, msg : %v",msg)
			http.Redirect(w, r, "/signup", 302)	
		}
		log.Printf("uuid:%s\n",  uuid)

		// store session info
		log.Println("sess start")
		session, _ := store.Get(r, "cookie-name")
		session.Values["authenticated"] = true
		session.Values["uuid"] = uuid
		session.Save(r, w)
		log.Println("sess saved")
		http.Redirect(w, r, "/upload", 302)	
	}
}


func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
        file, err := os.Open(path)
        if err != nil {
                return nil, err
        }
        defer file.Close()

        body := &bytes.Buffer{}
        writer := multipart.NewWriter(body)
        part, err := writer.CreateFormFile(paramName, filepath.Base(path))
        if err != nil {
                return nil, err
        }
        _, err = io.Copy(part, file)

        for key, val := range params {
			_ = writer.WriteField(key, val)
        }
        err = writer.Close()
        if err != nil {
            return nil, err
        }

        req, err := http.NewRequest("POST", uri, body)
        req.Header.Set("Content-Type", writer.FormDataContentType())
        return req, err
}

func upload_help ( photoRelativePath string)  string {// upload a local file to photo server  alejandroseaah.com/upload, and return photo id
        extraParams := map[string]string{
                "title":       "My Document",
                "author":      "Matt Aimonetti",
                "description": "A document with all the Go programming language secrets",
        }
        //request, err := newfileUploadRequest("http://alejandroseaah.com:4869/upload", extraParams, "file", photoRelativePath)
        request, err := newfileUploadRequest(conf["image_upload_url"].(string), extraParams, "file", photoRelativePath)
        if err != nil {
            log.Fatal(err)
			return "NULL"
        }
        client := &http.Client{}
        resp, err := client.Do(request)
        if err != nil {
            log.Fatal(err)
			return "NULL"
        } else {
			body := &bytes.Buffer{}
			_, err := body.ReadFrom(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			resp.Body.Close()
			strs := strings.Split(body.String() ,"http://yourhostname:4869/")
			strs1 := strings.Split(strs[1],"</a>")
			return strs1[0]  //photoID
    }
}


func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	session.Values["authenticated"] = false
	session.Values["uuid"] = ""
	session.Save(r,w)
	http.Redirect(w,r,"/login",302)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	//check auth
	if auth, ok := session.Values["authenticated"].(bool); !auth {
		_ = ok
		http.Error(w, "Forbidden", http.StatusForbidden)
		http.Redirect(w, r, "/signup", 302)
	}

	uuid := session.Values["uuid"].(string)	
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("tpl/upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			log.Println(err)
			return
		}
		defer file.Close()

		info := handler.Header["Content-Disposition"]
		info1 := strings.Split(strings.Join(info,""),"filename=\"")
		info2 := strings.Split(info1[1],"\"")
		uploadedFileName := info2[0]

		f, err := os.OpenFile( conf["tmp_file_dir"].(string) + "/" + handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
		localFile := conf["tmp_file_dir"].(string) + "/" + uploadedFileName
		photoID := upload_help(localFile)

		// delete file
		if photoID != "NULL" {
			session.Values["photoid"] = photoID
			session.Save(r, w)
			log.Println("saved")
			//update db	
			var reply string
			client, err := rpc.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
			if err != nil {
				log.Println(err)
			}

			args := util.Args2{ uuid, photoID}
			err = client.Call("Query.InitAvatar", args, &reply)
			client.Close()
			if err != nil {
				log.Fatal(err)
			}
			log.Println("reply: " + reply)
			byt := []byte(reply)
			var dat map[string]interface{}
			if err := json.Unmarshal(byt, &dat); err != nil {
				panic(err)
			}
			_ = err
			code := dat["code"].(float64)
			log.Println(code)
			if code != 0 {
				log.Println("failed to upload db")
				http.Redirect(w, r, "/upload", 302)
			}
			//update redis

			os.Remove(localFile)
			http.Redirect(w, r, "/home", 302)
		}else{//
			http.Redirect(w, r, "/upload", 302)
		}
	}
}


func editHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	//check auth
	if auth, ok := session.Values["authenticated"].(bool); !auth {
		_ = ok
		http.Error(w, "Forbidden", http.StatusForbidden)
		http.Redirect(w, r, "/signup", 302)
	}

	uuid := session.Values["uuid"].(string)	
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("tpl/edit.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseForm()
		var reply string
		nickname := strings.Join(r.Form["nname"],"")
		log.Println("nn " + nickname)
		if len(nickname) > 0 {
			client, err := rpc.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
			_ = err
			args := util.Args2{ uuid, nickname}
			err = client.Call("Query.ChangeNickname", args, &reply)
			client.Close()
			_ = reply
		}
		
		http.Redirect(w, r, "/home", 302)
	}
}




func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !auth {
		_ = ok
		log.Println("Log in first")
		http.Redirect(w, r, "/login", 302) // redirect to home page
	}

	uuid := session.Values["uuid"].(string)
	var reply string
	log.Println("uuid in sess :" + uuid )
	client, err := rpc.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	args := util.Args2{uuid,""}
	err = client.Call("Query.Lookup", args, &reply)
	log.Println("lookup: " + reply)
	client.Close()
	byt := []byte(reply)
	var dat map[string]interface{}
	json.Unmarshal(byt, &dat)
	code := dat["code"].(float64)
	pid := dat["photoid"].(string)
	nickname := dat["nickname"].(string)
	//avatar_url := "http://alejandroseaah.com:4869/"+ pid +"?w=600&h=600"
	avatar_url := conf["image_fetch_prefix"].(string) + "/"+ pid +"?w=600&h=600"
	_ = err
	_ = code
	data := HomeInfo{
		AvatarUrl :  avatar_url,
		Nickname : nickname,
	}

	t := template.Must(template.ParseFiles("tpl/home.html"))
	t.Execute(w, data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); auth {
		log.Println("Already logged in ")
		_ = ok
		http.Redirect(w, r, "/home", 302) // redirect to home page
	}

	if r.Method == "GET" {
		t, _ := template.ParseFiles("tpl/login.gtpl")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		// logic part of log in
		username := strings.Join(r.Form["username"],"")
		pwd := strings.Join(r.Form["password"],"")

		client, err := rpc.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
		if err != nil {
			log.Fatal("dialing:", err)
		}
		args := util.Args2{username,pwd}
		var reply string
		err = client.Call("Query.SignIn", args, &reply)
		client.Close()

		if err != nil {
			log.Fatal("error:", err)
		}
		log.Printf("response:%s\n",  reply)
		byt := []byte(reply)
		var dat map[string]interface{}
		if err := json.Unmarshal(byt, &dat); err != nil {
			panic(err)
		}
		code := dat["code"].(float64)
		uuid := dat["uuid"].(string)
		log.Println(code)
		if code != 0 {
			log.Println("failed to login")
			http.Redirect(w,r,"/login", 302)
		}
		log.Println("check login uuid: %s",uuid)
		//communicate with tcp server and proxy server
		session.Values["authenticated"] = true
		session.Values["uuid"] = uuid
		session.Save(r, w)
		http.Redirect(w, r, "/home", 302)
  }
}

var conf = make(map[string]interface{})

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	conf = util.ConfReader(dir + "/../../conf/setting.conf")
	logDir := conf["log_file_dir"].(string)

	f, err := os.OpenFile( dir + "/" + logDir + "/web_server.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: ", err)
	}
	defer f.Close()
	log.SetOutput(f)

	http.HandleFunc("/signup", signup)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/edit", editHandler)

	web_host := conf["web_server_host"].(string)
	web_port := conf["web_server_port"].(string)
	addr := web_host+":"+web_port
	log.Println("listening to addr  " +  addr)
	err = http.ListenAndServe(string(addr), nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

