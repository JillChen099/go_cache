package main

import (
	"go_cache/cache"
	"time"
	"fmt"
)

func main(){

	defaultExpiration,_ := time.ParseDuration("0.5h")
	gcInterval,_ := time.ParseDuration("3s")
	c := cache.NewCache(defaultExpiration,gcInterval)
	k1 := "hello Mr.Chen"
	expiration,_ := time.ParseDuration("5s")
	c.Set("k1",k1,expiration)
	s,_ := time.ParseDuration("1s")
	for {
		time.Sleep(s)
		if v, found := c.Get("k1"); found {
			fmt.Println("Found k1:", v)

		} else {
			fmt.Println("Not found k1")
			break
		}
	}
}