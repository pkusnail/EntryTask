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
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "9999"
	CONN_TYPE = "tcp"
)

type Args struct {
    A, B string
}

type Args2 struct {
        A,B string
}

type Args3 struct {
        A,B,C string
}


type Args4 struct {
        A,B,C,D string
}

type Register struct {
	realname string
	nickname string
	info string
}

type Resp struct {
	code int
	msg	string
	data [] string
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



func sayhelloName(w http.ResponseWriter, r *http.Request) {
    r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
    // attention: If you do not call ParseForm method, the following data can not be obtained form
    fmt.Println(r.Form) // print information on server side.
    fmt.Println("path", r.URL.Path)
    fmt.Println("scheme", r.URL.Scheme)
    fmt.Println(r.Form["url_long"])
    for k, v := range r.Form {
        fmt.Println("key:", k)
        fmt.Println("val:", strings.Join(v, ""))
    }
    fmt.Fprintf(w, "Hello astaxie!") // write data to response
}



func signup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		var rname, nname, info  []string
		var ok bool
		pnum := len(r.URL.Query())
		if pnum > 0{ 
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
		log.Println(popup)
		log.Println(rn)
		log.Println(nn)
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
		fmt.Println("real name:", realname)
		fmt.Println("nick name:", nickname)
		fmt.Println("pwd1:", r.Form["pwd1"])
		fmt.Println("pwd2:", r.Form["pwd2"])
		pwd1 := strings.Join(r.Form["pwd1"],"")
		pwd2 := strings.Join(r.Form["pwd2"],"")
		fmt.Println("pwd1:", pwd1)
		fmt.Println("pwd2:", pwd2)
		if pwd1 != pwd2 {
			http.Redirect(w, r, "/signup", 301)
		}
		//communicate with tcp server and proxy server  
		//client, err := rpc.Dial("tcp", "localhost:9999")
		client, err := rpc.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
		if err != nil {
			fmt.Println(err)
			log.Fatal("dialing:", err)
		}
		args := Args4{realname,nickname,pwd1,""}
		var reply string
		err = client.Call("Query.SignUp", args, &reply)
		client.Close()
		if err != nil {
			fmt.Println(err)
			log.Fatal("Query.SignUp error:", err)
		}
		fmt.Println("check : %v",reply)	
	byt := []byte(reply)
	var dat map[string]interface{}
	if err := json.Unmarshal(byt, &dat); err != nil {
		fmt.Println("careful :" + reply)
		panic(err)
	}
		code := dat["code"].(float64)
		uuid := dat["uuid"].(string)
		msg := dat["msg"].(string)
		if code != 0 {
			fmt.Println("Sign up failed, msg : %v",msg)
			http.Redirect(w, r, "/signup", 302)	
		}
		fmt.Printf("uuid:%s\n",  uuid)

		// store session info
		fmt.Println("sess start")
		session, _ := store.Get(r, "cookie-name")
		session.Values["authenticated"] = true
		session.Values["uuid"] = uuid
		session.Save(r, w)
		fmt.Println("sess saved")
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
        //path, _ := os.Getwd()
        //path += "/test.pdf"
        extraParams := map[string]string{
                "title":       "My Document",
                "author":      "Matt Aimonetti",
                "description": "A document with all the Go programming language secrets",
        }
        //request, err := newfileUploadRequest("http://alejandroseaah.com:4869/upload", extraParams, "file", "./shell.png")
        request, err := newfileUploadRequest("http://alejandroseaah.com:4869/upload", extraParams, "file", photoRelativePath)
        if err != nil {
            log.Fatal(err)
			return "NULL"
        }
        client := &http.Client{}
        resp, err := client.Do(request)
		//client.Close()
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
			fmt.Println(resp.StatusCode)
			fmt.Println(resp.Header)

			//fmt.Println(body)
			strs := strings.Split(body.String() ,"http://yourhostname:4869/")
			strs1 := strings.Split(strs[1],"</a>")
			fmt.Println(strs1[0])
			//defer client.Close()
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

	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("./tpl/upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		info := handler.Header["Content-Disposition"]
		info1 := strings.Split(strings.Join(info,""),"filename=\"")
		info2 := strings.Split(info1[1],"\"")
		uploadedFileName := info2[0]
		fmt.Fprintf(w, "%v", uploadedFileName )
		fmt.Fprintf(w, "%v", handler.Header)

		f, err := os.OpenFile("./tmp/" + handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
		localFile := "./tmp/" + uploadedFileName
		photoID := upload_help(localFile)
		fmt.Printf("see photo ID : %s", photoID)
		fmt.Fprintf(w, "photo ID : %v", photoID)

		// delete file
		if photoID != "NULL" {
			fmt.Println("good ,go on")
			//var err = os.Remove(localFile)
			//if isError(err) { return }
			session.Values["photoid"] = photoID
			session.Save(r, w)
			fmt.Println("saved")
			//update db	
			var reply string
			//client, err := rpc.Dial("tcp", "127.0.0.1:9999")
			client, err := rpc.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
			if err != nil {
				fmt.Println(err)
			}

			args := Args2{ uuid, photoID}
			err = client.Call("Query.InitAvatar", args, &reply)
			client.Close()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("reply: " + reply)
			byt := []byte(reply)
			var dat map[string]interface{}
			if err := json.Unmarshal(byt, &dat); err != nil {
				panic(err)
			}
			_ = err
			code := dat["code"].(float64)
			//uuid := dat["uuid"].(string)
			fmt.Println(code)
			if code != 0 {
				fmt.Println("failed to upload db")
				http.Redirect(w, r, "/upload", 302)
				//http.Redirect(w,r,"/login", 302)
			}
			//update redis

			fmt.Println("good")
			os.Remove(localFile)
			http.Redirect(w, r, "/home", 302)
		}else{//
			fmt.Println("wrong")
			http.Redirect(w, r, "/upload", 302)
		}
	}
}



func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !auth {
		_ = ok
		fmt.Println("Log in first")
		http.Redirect(w, r, "/login", 302) // redirect to home page
		//http.Error(w, "Forbidden", http.StatusForbidden)
		//return
	}

	//pid := session.Values["photoid"].(string)
	uuid := session.Values["uuid"].(string)
	var reply string

	//client, err := rpc.Dial("tcp", "localhost:9999")
	client, err := rpc.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	args := Args2{uuid,""}
	err = client.Call("Query.Lookup", args, &reply)
	fmt.Println("lookup: " + reply)
	client.Close()
	byt := []byte(reply)
	var dat map[string]interface{}
	json.Unmarshal(byt, &dat)
	code := dat["code"].(float64)
	pid := dat["photoid"].(string)
	nickname := dat["nickname"].(string)
	avatar_url := "http://alejandroseaah.com:4869/"+ pid +"?w=600&h=600"
	_ = err
	_ = code
	data := HomeInfo{
		//AvatarUrl : "\"" + avatar_url + "\"",
		AvatarUrl :  avatar_url,
		Nickname : nickname,
	}

	t := template.Must(template.ParseFiles("tpl/home.html"))
	t.Execute(w, data)

}


//todo

func editHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !auth {
		fmt.Println("Login in first ")
		_ = ok
		http.Redirect(w, r, "/login", 302) // redirect to home page
		//http.Error(w, "Forbidden", http.StatusForbidden)
		//return
	}

	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("tpl/edit.gtpl")
		t.Execute(w, nil)
	} else {

	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); auth {
		fmt.Println("Already logged in ")
		_ = ok
		http.Redirect(w, r, "/home", 302) // redirect to home page
		//http.Error(w, "Forbidden", http.StatusForbidden)
		//return
	}

	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
	t, _ := template.ParseFiles("tpl/login.gtpl")
	t.Execute(w, nil)
	} else {
	r.ParseForm()
	// logic part of log in
	username := strings.Join(r.Form["username"],"")
	pwd := strings.Join(r.Form["password"],"")

	fmt.Println("username:", username)
	fmt.Println("password:", pwd)


	//client, err := rpc.Dial("tcp", "localhost:9999")
	client, err := rpc.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	args := Args2{username,pwd}
	var reply string
	err = client.Call("Query.SignIn", args, &reply)
	client.Close()

	if err != nil {
		log.Fatal("error:", err)
	}
	fmt.Printf("response:%s\n",  reply)
/*	
	byt, err := json.Marshal(reply)
	if err != nil {
		panic(err)
	}
*/
	byt := []byte(reply)
	var dat map[string]interface{}
	if err := json.Unmarshal(byt, &dat); err != nil {
		panic(err)
	}
	code := dat["code"].(float64)
	uuid := dat["uuid"].(string)
	fmt.Println(code)
	if code != 0 {
		fmt.Println("failed to login")
		http.Redirect(w,r,"/login", 302)
	}
	fmt.Println("uuid: %s",uuid)
	//fmt.Println(dat)	
	//communicate with tcp server and proxy server
	session.Values["authenticated"] = true
	session.Values["uuid"] = uuid
	session.Save(r, w)
	http.Redirect(w, r, "/home", 302)
  }
}






func main() {
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/edit", editHandler)


	err := http.ListenAndServe(":9090", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

