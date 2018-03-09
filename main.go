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
	expiration,_ := time.ParseDuration("60s")
	c.Set("k1",k1,expiration)

	s,_ := time.ParseDuration("1s")  //验证缓存系统自动剔除过期key
	for {
		time.Sleep(s)
		if v, found := c.Get("k1"); found {
			fmt.Println("Found k1:", v)

		} else {
			fmt.Println("Not found k1")
			break
		}
	}



	k2 := "baby my love"
	c.Set("k2",k2,expiration)
	err := c.SaveToFile("C:/test")  //缓存持久化
	if err !=nil {
		fmt.Println(err)
	}


	err = c.LoadFile("C:/test") //读取缓存文件
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(c.Count())
	fmt.Println(c.Get("k2"))
}