package app

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

// Health is used to handle HTTP Health requests to this service. Use this for liveness
// probes or any other checks which only validate if the services is running.
func (srv *Server) Health(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.WriteHeader(http.StatusOK)
}

// Ready is used to handle HTTP Ready requests to this service. Use this for readiness
// probes or any checks that validate the service is ready to accept traffic.
func (srv *Server) Ready(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check other stuff here like DB connectivity, health of dependent services, etc.
	err := srv.kv.HealthCheck()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// middleware is used to intercept incoming HTTP calls and apply general functions upon
// them. e.g. Metrics, Logging...
func (srv *Server) middleware(n httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Log the basics
		srv.log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("HTTP Request to %s", r.URL)

		// Call registered handler
		n(w, r, ps)
	}
}

// Hello will handle any requests to /hello with a greating.
func (srv *Server) Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// Fetch Custom greeting from the DB
	g, err := srv.kv.Get("greeting")
	if err != nil {
		srv.log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Could not fetch data from database - %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	// Print greeting or say hello
	if len(g) > 0 {
		fmt.Fprintf(w, "%s", g)
		return
	}
	fmt.Fprintf(w, "%s", "Hello World")
}

// SetHello will handle any update requests to /hello to store our greating.
func (srv *Server) SetHello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		srv.log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Error reading body from request to %s", r.URL)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = srv.kv.Set("greeting", body)
	if err != nil {
		srv.log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Could not update database - %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", "Success")
}
