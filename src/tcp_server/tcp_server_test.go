package main

import "testing"
import "fmt"
import "encoding/json"

func Test_Nickname(t *testing.T) {
	rep := insertUser("rn1","nn1","pwd1","av1")
	byt := []byte(rep)
	var dat map[string]interface{}
	if err := json.Unmarshal(byt, &dat); err != nil {
		t.Log("careful :" + rep)
		t.Errorf("fail")
	}
	code := dat["code"].(float64)
	uuid := dat["uuid"].(string)
	msg := dat["msg"].(string)
	if code != 0 {
		t.Errorf("insert user  failed, msg : "+msg)
	}

	//updateNickname
}

func TestInsertAvatar(t *testing.T) {
	insertAvatar("testKey", "testValue")
	v := lookupAvatar("testKey")
	fmt.Println(v)
	if v != "testValue" {
		 t.Errorf("failed")
	}else{
		 t.Log("good")
	}
}
