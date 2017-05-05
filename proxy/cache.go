package proxy

import (
	"bytes"
	"httpproxy/cache"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var cacheBox *cache.CacheBox

func RegisterCacheBox(c *cache.CacheBox) {
	cacheBox = c
}

func CacheHandler(rw http.ResponseWriter, req *http.Request) {
	url := req.RequestURI
	log.Printf("Cache: TRY %v %v \n", req.Method, url)

	c := cacheBox.Get(url)

	if c != nil {
		if c.Verify() {
			log.Printf("===============\nCache: Get cache of %s \n============\n", url)
			c.WriteTo(rw)
			return
		}
		log.Printf("*******\nCache %s unverify\n*******\n", url)
		cacheBox.Delete(url)
	}

	RemoveHopByHopHeaders(req)
	tr := &http.Transport{Proxy: http.ProxyFromEnvironment}
	resp, err := tr.RoundTrip(req)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	cresp := new(http.Response)
	*cresp = *resp
	CopyResponse(cresp, resp)

	log.Printf("/////////\nCheck and store cache of %s\n////////////\n", url)
	go cacheBox.CheckAndStore(url, cresp)

	ClearHeaders(rw.Header())
	CopyHeaders(rw.Header(), resp.Header)
	rw.WriteHeader(resp.StatusCode)

	_, err = io.Copy(rw, resp.Body)
	if err != nil && err != io.EOF {
		log.Printf("Got an error when copy remote response to client\n")
	}
}

func CopyResponse(dest *http.Response, src *http.Response) {

	*dest = *src
	var bodyBytes []byte

	if src.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(src.Body)
	}

	// Restore the io.ReadCloser to its original state
	src.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	dest.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
}
