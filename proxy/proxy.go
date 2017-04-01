package proxy

import (
	"net/http"
	"strings"
	"fmt"
	"io/ioutil"
	"io"
	"log"
)

func NewProxy(addr string) {
	log.SetFlags(log.LstdFlags|log.Lshortfile)
	http.HandleFunc("/",handle)
	log.Println("Start listen port ", addr)
	err := http.ListenAndServe(addr,nil)
	if err != nil {
		log.Fatal("ListenAndServe: ",err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	if ip := strings.Split(r.RemoteAddr, ":")[0]; ip == "127.1.0.1" {
		fmt.Println("Block ip : ", ip)
		return
	}
	switch r.Host {
	case "www.jianshu.com":
		http.Redirect(w,r,"http://www.baidu.com",301)
	case "www.taobao.com":
		w.Write([]byte("<h1>垃圾淘宝，毁我青春，耗我钱财，颓我精神</h1>"))
	default:
		// RequestURI is the unmodified Request-URI of the
		// Request-Line (RFC 2616, Section 5.1) as sent by the client
		// to a server. Usually the URL field should be used instead.
		// It is an error to set this field in an HTTP client request.
		//
		r.RequestURI = ""
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Add(k,vv)
			}
		}
		for _, c := range resp.Cookies() {
			w.Header().Add("set-Cookie", c.Raw)
		}
		w.WriteHeader(resp.StatusCode)
		result, err := ioutil.ReadAll(resp.Body)
		if err != nil && err != io.EOF {
			panic(err)
		}
		w.Write(result)
	}
}
