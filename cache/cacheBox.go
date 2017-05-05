package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	db "github.com/syndtr/goleveldb/leveldb"
	"log"
	"net/http"
)

type CacheBox struct {
	pool *db.DB
}

func MD5Url(url string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(url)))
}

func NewCacheBox(path string) *CacheBox {
	pool, err := db.OpenFile(path, nil)
	if err != nil {
		return nil
	}
	return &CacheBox{pool: pool}
}

func (c *CacheBox) Get(url string) *Cache {
	if cache := c.get(MD5Url(url)); cache != nil {
		log.Println("GET CACHE!!!!!!")
		return cache
	}
	//	log.Println("NO SUCH CACHE")
	return nil
}

func (c *CacheBox) get(md5Url string) *Cache {
	b, err := c.pool.Get([]byte(md5Url), nil)
	if err != nil {
		return nil
	}
	cache := new(Cache)
	json.Unmarshal(b, &cache)
	return cache
}

func (c *CacheBox) Delete(url string) {
	c.delete(MD5Url(url))
}
func (c *CacheBox) delete(md5Url string) {
	err := c.pool.Delete([]byte(md5Url), nil)
	if err != nil {
		log.Panic(err)
	}
}

func (c *CacheBox) CheckAndStore(url string, response *http.Response) {
	if !IsCache(response) {
		log.Println("This not allowed cache")
		return
	}

	cache := New(response)
	if cache == nil {
		return
	}
	cache.URL = url

	log.Println("&&&&&&&&&STORE CACHE&&&&&&&", url)

	md5Url := MD5Url(url)
	b, err := json.Marshal(cache)
	if err != nil {
		log.Println(err)
		return
	}
	err = c.pool.Put([]byte(md5Url), b, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("&&&&&&&&&&&SUCCESS STORE&&&&&&&&&&")
	return
}

func (c *CacheBox) Close() {
	c.pool.Close()
}
