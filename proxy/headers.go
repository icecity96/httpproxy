package proxy

import "net/http"

// Hop-by-hop headers, which are meaningful only for a single transport-level connection
// and are not stored by caches or forwarded by proxies.
func RemoveHopByHopHeaders(req *http.Request) {
	req.RequestURI = ""
	req.Header.Del("Proxy-Connection")
	req.Header.Del("Connection")
	req.Header.Del("Keep-Alive")
	req.Header.Del("Proxy-Authenticate")
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("TE")
	req.Header.Del("Trailers")
	req.Header.Del("Transfer-Encoding")
	req.Header.Del("Upgrade")
}

func CopyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func ClearHeaders(headers http.Header) {
	for key, _ := range headers {
		headers.Del(key)
	}
}
