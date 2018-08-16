package util

import (
	"redis"
	"log"
	"bufio"
	"strings"
	"os"
	"io"
)

func RedisPut( key string, val string) bool {
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

func RedisGet( key string) string {
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
	return string(value[:])
}

func ConfReader(path string) map[string]interface{} {
	var conf=make( map[string]interface{} )
	f, _ := os.Open(path)
	buf := bufio.NewReader(f)
	for {
		l, err := buf.ReadString('\n')
		line := strings.TrimSpace(l)
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			if len(line) == 0 {
				break
			}
		}
		switch {
		case len(line) == 0:
			case line[0] == '[' && line[len(line)-1] == ']': 
		//session  "[db]"
			section := strings.TrimSpace(line[1 : len(line)-1])
			_ = section
		default:
			//dnusername = xiaowei 这种的可以匹配存储
			i := strings.IndexAny(line, "=")
			conf[strings.TrimSpace(line[0:i])] = strings.TrimSpace(line[i+1:])

		}
	}
	log.Println("Check configuration : ")
	for k, v := range conf {
		log.Println(k," => ", v)
	}

	return conf
}
