package app

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

// server is used as an interface for managing the HTTP server.
type server struct {
	// httpServer is the primary HTTP server.
	httpServer *http.Server

	// httpRouter is used to store and access the HTTP Request Router.
	httpRouter *httprouter.Router
}

// Health is used to handle HTTP Health requests to this service.
func (s *server) Health(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.WriteHeader(http.StatusOK)
}

// Ready is used to handle HTTP Ready requests to this service.
func (s *server) Ready(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check other stuff here like DB connectivity, health of dependent services, etc.
	err := db.HealthCheck()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// middleware is used to intercept incoming HTTP calls and apply general functions upon
// them. e.g. Metrics, Logging...
func (s *server) middleware(n httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Log the basics
		log.WithFields(logrus.Fields{
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
func (s *server) Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// We will always succeed in saying hello
	w.WriteHeader(http.StatusOK)

	// Fetch Custom greeting from the DB
	g, err := db.Get("greeting")
	if err != nil {
		log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Could not fetch data from database - %s", err)
	}

	// Print greeting or say hello
	if len(g) > 0 {
		fmt.Fprintf(w, "%s", g)
		return
	}
	fmt.Fprintf(w, "%s", "Hello World")
}

// SetHello will handle any update requests to /hello to store our greating.
func (s *server) SetHello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Error reading body from request to %s", r.URL)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = db.Set("greeting", body)
	if err != nil {
		log.WithFields(logrus.Fields{
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
	fmt.Fprint(w, "%s", "Success")
}
