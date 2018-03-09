# Dockerfile Health Check Example

This repository holds a very basic Go application that acts as an HTTPS frontend for Redis GET (GET) & SET (POST/PUT) commands. This Go application also implements a `/status` end point that returns `200` if Redis is up and accessible. The `Dockerfile` in this repository uses Health Checks to monitor this end point to see if the Go application is in fact "healthy". 

This example is the source used within an upcoming CodeShip article about Docker Health Checks.
