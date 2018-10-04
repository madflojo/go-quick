# Health Checks Example

This repository holds a very basic Go application that acts as an HTTPS frontend for Redis GET (GET) & SET (POST/PUT) 
commands available via the `/kv` end point.

This simple application also implements a `/healthy` end point that returns a `200`, to be an example of "Liveness 
Probes". To show "Readiness Probes" there is a `/ready` end point that only returns a `200` if Redis is up and 
accessible.

While this code is functioning, it is not meant to be used for anything more than an example. This exists as a example 
for future Articles and Talks about health checks, readiness, and graceful shutdowns.
