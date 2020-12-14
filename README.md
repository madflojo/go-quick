# Go Quick

This project is a template repository that contains an example HTTPS Hello World API written in Go. The goal of this project is to help new developers learn how to structure Go applications by example. As well as provide a production-ready base application that anyone can fork and run.

The main features of this template are:

- Liveness probe support via `/health` end-point
- Readiness probe support via `/ready` end-point
- Graceful shutdown with a SIGTERM signal trap
- Modular Key-Value Database integration using <https://github.com/madflojo/hord>

TODO:

- Add basic metrics
- Add tracing via OpenTracing
