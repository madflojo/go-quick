// This is an example application to show the power of Dockerfile Health Checks
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

func main() {
	// Connect to Redis
	log.Print("INFO: Connecting to Redis")
	c, err := redis.Dial("tcp", "redis:6379")
	if err != nil {
		log.Printf("CRITICAL: Could not connect to redis - %s", err)
	}
	defer c.Close()
	rConn = c

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

	}

	// Send data back to client
	_, err = ctx.Write(rsp)
	if err != nil {
		log.Printf("INFO: Could not write response on connection - %d", ctx.ID())
	}
}
