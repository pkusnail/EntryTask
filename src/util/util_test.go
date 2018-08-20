package util

import "testing"
import "os"
import "log"
import "fmt"
import "strconv"

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
	if conf["tcp_server_host"] != "127.0.0.1" {
		t.Errorf("failed")
	}

}
func TestRedis(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var conf map[string]interface{}
	conf = ConfReader(dir + "/../../conf/setting.conf")
	redisPool := &RedisPool{}
	redisHost := conf["redis_host"].(string)
	redisPort := conf["redis_port"].(string)
	redisAddr := redisHost + ":" + redisPort
	maxIdle, err := strconv.Atoi(conf["redis_max_idle"].(string))
	maxActive, err := strconv.Atoi(conf["redis_max_active"].(string))
	log.Println("redis addr : " + redisAddr)
	//initRedis("tcp", redisAddr)
	redisPool = &RedisPool{
		MaxIdle : &maxIdle,
		MaxActive : &maxActive,
		Addr : &redisAddr,
		Pool : nil,
	}
	redisPool.NewPool()
	redisPool.RedisSet("testKey", "testValue")
	v := redisPool.RedisGet("testKey")
	fmt.Println(v)
	if v != "testValue" {
		 t.Errorf("failed")
	}else{
		 t.Log("good")
	}
}
