package webserver

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"
)

type StatsWriter struct {
	w      http.ResponseWriter
	Status int
	Length int
	Start  time.Time
}

type StatsConnWriter struct {
	net.Conn
	sw *StatsWriter
}

func NewStatsWriter(w http.ResponseWriter) *StatsWriter {
	return &StatsWriter{w: w, Start: time.Now(), Status: 200}
}

func (w *StatsWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.w.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("No Hijacker!")

	}

	c, b, err := hj.Hijack()
	return StatsConnWriter{c, w}, b, err
}

func (w *StatsWriter) Header() http.Header {
	return w.w.Header()
}

func (w *StatsWriter) Write(b []byte) (int, error) {
	w.Length += len(b)
	return w.w.Write(b)
}

func (w *StatsWriter) WriteHeader(code int) {
	w.Status = code
	w.w.WriteHeader(code)
}

func (n StatsConnWriter) Write(b []byte) (int, error) {
	n.sw.Length += len(b)
	return n.Conn.Write(b)
}
