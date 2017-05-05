package cache

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// MORE INFO: https://my.oschina.net/leejun2005/blog/369148
type Cache struct {
	Header       http.Header `json:"header"`
	Body         []byte      `json:"body"`
	StatusCode   int         `json:"status_code"`
	URL          string      `json:"url"`
	LastModified string      `json:"last_modified"`
	ETag         string      `json:"etag"`
	MustVerified bool        `json:"must_verified"`
	Vlidity      time.Time   `json:"vlidity"`
	maxAge       int64       `json:"-"`
}

func New(resp *http.Response) *Cache {
	c := new(Cache)
	c.Header = make(http.Header)
	CopyHeaders(c.Header, resp.Header)

	c.StatusCode = resp.StatusCode

	var err error
	c.Body, err = ioutil.ReadAll(resp.Body)
	if err != nil && err != io.EOF {
		log.Println(err)
	}

	if c.Header == nil {
		return nil
	}

	c.ETag = c.Header.Get("ETag")
	c.LastModified = c.Header.Get("Last-Modified")

	cacheControl := c.Header.Get("Cache-Control")

	if strings.Index(cacheControl, "no-cache") != -1 ||
		strings.Index(cacheControl, "must-revalidate") != -1 ||
		strings.Index(cacheControl, "proxy-revalidate") != -1 {
		c.MustVerified = true
		return nil
	}

	// Expires策略 其优先级别较低
	if expires := c.Header.Get("Expires"); expires != "" {
		c.Vlidity, err = time.Parse(http.TimeFormat, expires)
		if err != nil {
			return nil
		}
	}

	if maxAge := getAge("max-age", cacheControl); maxAge != -1 {
		var t time.Time

		if date := c.Header.Get("Date"); date != "" {
			t, err = time.Parse(time.RFC1123, date)
			if err != nil {
				return nil
			}
		} else {
			t = time.Now().UTC()
		}

		c.Vlidity = t.Add(time.Duration(maxAge) * time.Second)
		c.maxAge = maxAge
	} else {
		// 默认最大1天
		c.maxAge = 24 * 60 * 60
	}
	return c
}

// 检查缓存是否有效
func (c *Cache) Verify() bool {
	if c.MustVerified == false && c.Vlidity.After(time.Now().UTC()) {
		return true
	}

	req, err := http.NewRequest("Get", c.URL, nil)
	if err != nil {
		return false
	}

	if c.LastModified != "" {
		req.Header.Add("If-Modified-Since", c.LastModified)
	}
	if c.ETag != "" {
		req.Header.Add("If-None-Match", c.ETag)
	}

	tr := &http.Transport{Proxy: http.ProxyFromEnvironment}
	resp, err := tr.RoundTrip(req)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusNotModified {
		return false
	}
	return true
}

func CopyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func getAge(pattern, src string) int64 {
	var num int64
	if index := strings.Index(src, pattern); index != -1 {
		for i := index + len(pattern) + 1; i < len(src); i++ {
			if '0' <= src[i] && src[i] <= '9' {
				num = num*10 + int64(src[i]-'0')
			}
		}
		return num
	}
	return -1
}

func (c *Cache) WriteTo(rw http.ResponseWriter) (int, error) {
	CopyHeaders(rw.Header(), c.Header)
	rw.WriteHeader(c.StatusCode)
	return rw.Write(c.Body)
}

//IsCache checks whether response can be stored as cache
func IsCache(resp *http.Response) bool {

	Cache_Control := resp.Header.Get("Cache-Control")
	Content_type := resp.Header.Get("Content-Type")
	if strings.Index(Cache_Control, "private") != -1 ||
		strings.Index(Cache_Control, "no-store") != -1 ||
		strings.Index(Content_type, "application") != -1 ||
		strings.Index(Content_type, "video") != -1 ||
		strings.Index(Content_type, "audio") != -1 ||
		(strings.Index(Cache_Control, "max-age") == -1 &&
			strings.Index(Cache_Control, "s-maxage") == -1 &&
			resp.Header.Get("Etag") == "" &&
			resp.Header.Get("Last-Modified") == "" &&
			(resp.Header.Get("Expires") == "" || resp.Header.Get("Expires") == "0")) {
		return false
	}
	return true
}
