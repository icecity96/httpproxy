package main

import (
	"httpproxy/proxy"
)

func main() {
	proxy.NewProxy(":8888")
}
