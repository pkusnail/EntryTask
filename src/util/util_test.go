package util

import "testing"

func TestRedis(t *testing.T) {
	RedisPut("testKey", "testValue")
	v := RedisGet("testKey")
	if v != "testValue" {
		 t.Errorf("failed")
	}else{
		 t.Log("good")
	}
}
