package util

import "testing"
import "os"
import "log"

func TestConfReader(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile( dir + "/../../log/web_server.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	conf := ConfReader(dir + "/../../conf/setting.conf")
	if conf["tcp_server_host"] != "localhost" {
		t.Errorf("failed")
	}

}
func TestRedis(t *testing.T) {
	RedisPut("testKey", "testValue")
	v := RedisGet("testKey")
	if v != "testValue" {
		 t.Errorf("failed")
	}else{
		 t.Log("good")
	}
}
