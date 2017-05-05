package proxy

import (
	"fmt"
	"httpproxy/cache"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func NewProxy(addr string) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	RegisterCacheBox(cache.NewCacheBox("./proxycache"))
	http.HandleFunc("/", handle)
	log.Println("Start listen port ", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	if ip := strings.Split(r.RemoteAddr, ":")[0]; ip == "127.1.0.1" {
		fmt.Println("Block ip : ", ip)
		return
	}
	switch r.Host {
	case "www.jianshu.com":
		http.Redirect(w, r, "http://www.baidu.com", 301)
	case "www.taobao.com":
		w.Write([]byte("<h1>垃圾淘宝，毁我青春，耗我钱财，颓我精神</h1>"))
	default:
		if r.Method == "CONNEcT" {
			HttpsHandler(w, r)
		} else if r.Method == "GET" {
			CacheHandler(w, r)
		} else {
			HttpHandler(w, r)
		}
	}
}

var HTTP_200 = []byte("HTTP/1.1 200 Connection Established\r\n\r\n")

// 自己接管连接
func HttpsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTPS: %v %v\n",r.Method, r.URL.String())
	hj, _ := w.(http.Hijacker)
	client, _, err := hj.Hijack() //获取客户端和代理服务器的tcp连接
	if err != nil {
		http.Error(w, "Failed", http.StatusBadRequest)
		return
	}

	remote, err := net.Dial("tcp", r.URL.Host) // 服务器与代理服务器的tcp连接
	if err != nil {
		client.Close()
		return
	}

	client.Write(HTTP_200)
	go copyRemoteToClient(remote, client)
	go copyRemoteToClient(client, remote)
}

func HttpHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("HTTP: %v %v \n", req.Method, req.URL.Host)

	RemoveHopByHopHeaders(req)

	tr := &http.Transport{Proxy: http.ProxyFromEnvironment}
	resp, err := tr.RoundTrip(req)
	if err != nil {
		log.Println("%v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	ClearHeaders(w.Header())
	CopyHeaders(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil && err != io.EOF {
		log.Println(" got an error when copy remote response to client.%v\n", err)
		return
	}
}

func copyRemoteToClient(remote, client net.Conn) {
	defer func() {
		remote.Close()
		client.Close()
	}()

	_, err := io.Copy(remote, client)
	if err != nil && err != io.EOF {
		return
	}
}
