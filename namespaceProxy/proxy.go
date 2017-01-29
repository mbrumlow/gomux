package namespaceProxy

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type NamespaceProxy struct {
	Namespace    string
	ReverseProxy *httputil.ReverseProxy
	Director     func(req *http.Request)
	ErrorLog     *log.Logger
	tlsConfig    *tls.Config
}

func NewNamespaceProxy(namespace string, target *url.URL) *NamespaceProxy {

	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = trimNamespace(target.Path, req.URL.Path, namespace)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
	}

	return &NamespaceProxy{
		Namespace:    namespace,
		ReverseProxy: &httputil.ReverseProxy{Director: director},
		Director:     director,
		tlsConfig:    &tls.Config{},
	}

}

func (p *NamespaceProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	// If it is not a websocket then use Go's native reverse proxy configured
	// with our custom director.
	if !isws(req) {
		p.ReverseProxy.ServeHTTP(rw, req)
		return
	}

	outreq := new(http.Request)
	*outreq = *req // shallow copy

	p.Director(outreq)
	host := outreq.URL.Host

	if !strings.Contains(host, ":") {
		if outreq.URL.Scheme == "wss" {
			host = host + ":443"
		} else {
			host = host + ":80"
		}
	}

	dial := func(network, address string) (net.Conn, error) {
		return net.Dial(network, address)
	}

	if outreq.URL.Scheme == "wss" {
		dial = func(network, address string) (net.Conn, error) {
			return tls.Dial("tcp", host, p.tlsConfig)
		}
	}

	d, err := dial("tcp", host)
	if err != nil {
		http.Error(rw, "Error forwarding request.", 500)
		p.logf("Error dialing websocket back end %s: %v", outreq.URL, err)
		return
	}

	hj, ok := rw.(http.Hijacker)
	if !ok {
		http.Error(rw, "No hijacker!", 500)
		return
	}

	nc, _, err := hj.Hijack()
	if err != nil {
		p.logf("Hijack error: %v", err)
		return
	}

	defer nc.Close()
	defer d.Close()

	err = outreq.Write(d)
	if err != nil {
		p.logf("Error copying request to target: %v", err)
		return
	}
	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}
	go cp(d, nc)
	go cp(nc, d)
	<-errc
}

func (p *NamespaceProxy) logf(format string, args ...interface{}) {
	if p.ErrorLog != nil {
		p.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func isws(r *http.Request) bool {

	if !strings.Contains(r.Header.Get("Upgrade"), "websocket") {
		return false
	}

	if !strings.Contains(r.Header.Get("Connection"), "Upgrade") {
		return false
	}

	return true
}

func trimNamespace(target, path, namespace string) string {
	return singleJoiningSlash(
		target,
		strings.TrimPrefix(path, "/"+namespace))
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
