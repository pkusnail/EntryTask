package main

import "testing"
import "fmt"
import "encoding/json"


func Test_Nickname(t *testing.T) {
	rep := insertUser("rn23","nn1","pwd1","av1")
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
		t.Errorf("insert user  failed, msg : " + msg)
	}
	_ =uuid
	updateNickname( uuid,"changedNN")
}

func TestInsertAvatar(t *testing.T) {
	insertAvatar("testKey", "testValue")
	v := lookupAvatar("testKey")
	//fmt.Println(v)
	//{"code":0,"msg":"success","photoid":"testValue"}
	if v != "{\"code\":0,\"msg\":\"success\",\"photoid\":\"testValue\"}" {
		fmt.Println(v)
		 t.Errorf("failed")
	}else{
		 t.Log("good")
	}
}
