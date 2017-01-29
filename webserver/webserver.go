package webserver

import (
	"net/http"
	"sync/atomic"
	"time"
)

type Server struct {
	Handler http.Handler

	ActiveRequest uint64
	TotalRequest  uint64
	Logger        WebLog

	// TODO(mbrumlow):
	//
	// - Stats per error code.
	// - Map of all active request mapped to their counts.
	// -- This should include duration, and bytes sent.
	// - Last n finished request (live configurable).
	// -- This should include URL, response code, duration, and bytes sent.
	//
	// - Logging to std out, or to a logger interface.
}

type WebLog interface {
	Log(end time.Time, w *StatsWriter, r *http.Request)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	sw := NewStatsWriter(w)
	defer s.log(sw, r)

	atomic.AddUint64(&s.ActiveRequest, 1)

	// Check URL for override API and configuration path.
	s.Handler.ServeHTTP(sw, r)

	atomic.AddUint64(&s.ActiveRequest, ^uint64(0))
	atomic.AddUint64(&s.TotalRequest, 1)
}

func (s *Server) log(sw *StatsWriter, r *http.Request) {
	if s.Logger != nil {
		s.Logger.Log(time.Now(), sw, r)
	}
}
