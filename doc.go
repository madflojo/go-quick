/*
Health Checks Example - In Go

This project is a base example for an HTTP Hello World service written in Go. What sets this example apart is that it emphasizes resiliency features.


Those features are as follows:  

- Liveness probe support via `/health` end-point 

- Readiness probe support via `/ready` end-point 

- Graceful shutdown with SIGTERM signal trap  

TODO:

- Add basic metrics 

- Add tracing via OpenTracing


If you are looking to start a simple Go HTTP service, you could fork this repository and start from there.

*/
package example
