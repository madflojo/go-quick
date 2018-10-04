// This is an example application to show the power of Health Checks
package main

import (
	"github.com/garyburd/redigo/redis"
	"github.com/valyala/fasthttp"
	"log"
	"regexp"
)

// Redis Connection
var rConn redis.Conn

// Regexs for our routes
var readyRoute = regexp.MustCompile(`\/ready.*`)
var healthyRoute = regexp.MustCompile(`\/healthy.*`)
var kvRoute = regexp.MustCompile(`\/kv.*`)

// Healthy indicators
var healthy bool
var ready bool

func main() {
	// Connect to Redis
	log.Print("INFO: Connecting to Redis")
	c, err := redis.Dial("tcp", "redis:6379")
	if err != nil {
		log.Printf("CRITICAL: Could not connect to redis - %s", err)
	}
	defer c.Close()
	rConn = c

	// Set healthy to true
	healthy = true
	ready = true

	// Start Fasthttp listener
	log.Print("INFO: Starting Fasthttp listener")
	err = fasthttp.ListenAndServeTLS("0.0.0.0:8443", "/etc/ssl/example.cert", "/etc/ssl/example.key", httpHandler)
	if err != nil {
		log.Fatalf("FATALITY: Could not start Fasthttp listener - %s", err)
	}
}

func httpHandler(ctx *fasthttp.RequestCtx) {
	var rsp []byte
	var err error

	// If request is to /health
	if healthyRoute.Match(ctx.Path()) {

		// Fail if healthy global is false
		if !healthy {
			log.Printf("CRITICAL: Application shows unhealthy status, returning 500 to liveness probe")
			ctx.Error("Application in unhealthy state", 500)
			return
		}

		// Send data back to client
		_, err = ctx.Write(rsp)
		if err != nil {
			log.Printf("INFO: Could not write response on connection - %d", ctx.ID())
		}
		return
	}

	// If request is to /ready
	if readyRoute.Match(ctx.Path()) {

		// Fail if ready global is false
		if !ready {
			log.Printf("WARNING: Readiness probe requested, application is not ready")
			ctx.Error("Application is not in ready state", 503)
			return
		}

		// Check Redis availability, and fail accordingly
		_, err = rConn.Do("ECHO", string("ping"))
		if err != nil {
			log.Printf("WARNING: Redis ping failed, reverting to non-ready state")
			ctx.Error("Application is not in ready state", 503)
			return
		}

		// Send data back to client
		_, err = ctx.Write(rsp)
		if err != nil {
			log.Printf("INFO: Could not write response on connection - %d", ctx.ID())
		}
		return
	}

	// If request is to /kv
	if kvRoute.Match(ctx.Path()) {

		// If GET retrieve key from Redis
		if ctx.IsGet() {
			rsp, err = redis.Bytes(rConn.Do("GET", string(ctx.Path())))
			if err != nil {
				log.Printf("INFO: Could not fetch data from Redis - %s", err)
				ctx.Error("Could not fetch key", 404)
				return
			}
		}

		// If POST insert key into Redis
		if ctx.IsPost() || ctx.IsPut() {
			_, err := rConn.Do("SET", string(ctx.Path()), string(ctx.PostBody()))
			if err != nil {
				log.Printf("INFO: Could not insert data into Redis - %s", err)
				ctx.Error("Could not insert key", 500)
				return
			}
		}

		// Send data back to client
		_, err = ctx.Write(rsp)
		if err != nil {
			log.Printf("INFO: Could not write response on connection - %d", ctx.ID())
		}
		return
	}
}
