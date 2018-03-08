package cache

import (
	"time"
	"sync"
	"fmt"
	"io"
	"encoding/gob"
	"os"
)

type Item struct {
	Data    interface{}
	Expiration	int64
}


func (item Item) Expired() bool {  //判断缓存过期时间

	if item.Expiration == 0 {

		return false
	}
	return time.Now().UnixNano() >item.Expiration

}

const (
	//没有过期时间的标志

	NoExpiration time.Duration = -1

	//默认的过期时间
	DefaultExpration time.Duration = 0


)

type Cache struct {

	defaultExpration 	time.Duration
	items				map[string]Item
	mu      			sync.RWMutex  //读写锁
	gcInterval			time.Duration // 过期数据项清理周期
	stopGc     			chan bool
	}

func (c *Cache) gcLoop() {  //通过time.Ticker 定期执行 DeleteExpired() 方法，从而清理过期的数据项
	ticker := time.NewTicker(c.gcInterval)
	for {
		select {
			case <- ticker.C:
				c.DeleteExpired()
			case <- c.stopGc:
				ticker.Stop()
				return
		}
	}

}

func (c *Cache) delete(k string) {
	delete(c.items,k)
}

func (c *Cache) DeleteExpired() { //删除过期数据

	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()

	for k,v := range c.items{
		if v.Expiration >0 && now >v.Expiration {
			c.delete(k)
		}
	}
}


func (c *Cache) Set(k string,v interface{},d time.Duration) { //设置数据项

	var e int64
	if d == DefaultExpration {
		d = c.defaultExpration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[k] = Item{
		v,
		e,
	}
}



func (c *Cache) set(k string,v interface{},d time.Duration) { //设置数据项

	var e int64
	if d == DefaultExpration {
		d = c.defaultExpration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.items[k] = Item{
		v,
		e,
	}
}

func (c *Cache) get(k string) (interface{}, bool) { //获取数据项,还需判断是否过期
	item,found := c.items[k];
	if !found {
		return nil,false
	}
	if item.Expired() {
		return nil,false
	}
	return item.Data,true

}


func (c *Cache) Add(k string,v interface{},d time.Duration)error{ //添加数据项，如果数据项已经存在，则返回错误
	c.mu.Lock()
	_,found := c.get(k)
	if found{
		c.mu.Unlock()
		return fmt.Errorf("Item %s already exists",k)
	}
	c.set(k,v,d)
	c.mu.Unlock()
	return nil

}

func (c *Cache) Get(k string) (interface{},bool) { //获取数据项
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.get(k)
}


func (c *Cache) Replace(k string,v interface{},d time.Duration) error{ //替换一个存在的数据项
	c.mu.Lock()
	_,found := c.get(k)
	if !found {
		c.mu.Unlock()
		return fmt.Errorf("Item %s doesnt exist",k)
	}
	c.set(k,v,d)
	c.mu.Unlock()
	return nil
}

func (c *Cache) Delete(k string) { //删除一个数据项

	c.mu.Lock()
	c.delete(k)
	c.mu.Unlock()

}


func (c *Cache) Save(w io.Writer) (err error) { // 将缓存数据写入到io.Writer中
	enc := gob.NewEncoder(w)
	defer func() {
		if x:=recover(); x !=nil {
			err = fmt.Errorf("Error registering item types with Gob library")


		}
	}()
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _,v := range c.items{
		gob.Register(v.Data)
	}
	err = enc.Encode((&c.items))
	return
}


// 保存数据项到文件中

func (c *Cache) SaveToFile (file string) error {
	f,err := os.Create(file)
	if err != nil {

		return err
	}
	if err = c.Save(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()

}

func (c *Cache) Load (r io.Reader) error {  //从io.reader 中读取数据项
	dec := gob.NewDecoder(r)
	items := map[string]Item{}
	err := dec.Decode(&items)
	if err == nil {
		c.mu.Lock()
		defer c.mu.Unlock()
		for k,v := range items{
			ov,found := c.items[k]
			if !found || ov.Expired(){
				c.items[k] = v
			}
		}

	}
	return err

}


func (c *Cache) LoadFile (file string) error { //从文件中加载缓存数据项

	f,err := os.Open(file)
	if err != nil {
		return err
	}
	if err = c.Load(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func (c *Cache) Count() int { //返回缓存数据项的数据量
	c.mu.RLock()
	defer  c.mu.RUnlock()
	return len(c.items)
}

func (c *Cache) Flush(){ //清空缓存
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = map[string]Item{}

}

//停止过期缓存清理开关

func (c *Cache)StopGc (){
	c.stopGc <- true
}


//创建一个缓存系统

func NewCache(defaultExpiration,gcInteval time.Duration) *Cache {
	c := &Cache{
		defaultExpration:	defaultExpiration,
		gcInterval:         gcInteval,
		items:              map[string]Item{},
		stopGc:             make(chan bool),
	}
	go c.gcLoop()
	return c

}










