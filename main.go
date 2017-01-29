package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"time"

	"github.com/mbrumlow/gomux/dispatch"
	"github.com/mbrumlow/gomux/webserver"
)

var httpPort = flag.Int("httpPort", 8080, "HTTP bind port")

type httpmux struct {
}

type SimpleLogger struct {
}

func main() {

	flag.Parse()

	d := dispatch.NewDispatch()
	ws := &webserver.Server{
		Handler: d,
		Logger:  &SimpleLogger{},
	}

	/*
		// Example namespace to secondary server.

		p := "/test1/test2/test3/test4/"
		u, _ := url.Parse("http://localhost:8081")
		r := namespaceProxy.NewNamespaceProxy(p, u)
		d.AddNamespace(p, r)
	*/

	// Namespace to default root handler.
	d.AddNamespace("/", &httpmux{})

	s := &http.Server{
		Addr:    ":8080",
		Handler: ws,
	}
	log.Fatal(s.ListenAndServe())

	/*
		// Code stub for let's encrypt.

		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("test2.brumlow.io"),
			Cache:      autocert.DirCache("cache"),
		}
		s := &http.Server{
			Addr:      "162.218.228.235:https",
			Handler:   &httpmux{},
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		}
		log.Fatal(s.ListenAndServeTLS("", ""))
	*/

}

func (h *httpmux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World, %q", html.EscapeString(r.URL.Path))
}

func (l *SimpleLogger) Log(end time.Time, sw *webserver.StatsWriter, r *http.Request) {
	fmt.Printf("%s - [%v] [%8s] \"%v %v %v\" %d %d \"%s\"\n",
		r.RemoteAddr,
		time.Now().Format("02/Jan/2006:15:04:05 -0700"),
		end.Truncate(time.Microsecond).Sub(sw.Start.Truncate(time.Microsecond)),
		r.Method,
		html.EscapeString(r.RequestURI),
		r.Proto,
		sw.Status,
		sw.Length,
		l.getHeader(r, "User-Agent"))
}

func (l *SimpleLogger) getHeader(r *http.Request, key string) string {
	v := r.Header.Get(key)
	if len(v) > 0 {
		return v
	}
	return "-"
}
