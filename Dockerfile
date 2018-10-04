FROM golang:latest

RUN go get github.com/valyala/fasthttp
RUN go get github.com/garyburd/redigo/redis

ADD . /go/src/github.com/madflojo/cloudy-healthchecks-example
RUN go install github.com/madflojo/cloudy-healthchecks-example

# Create a sample SSL Cert
RUN openssl genrsa -out /etc/ssl/example.key 4096 && \
    openssl req -x509 -new -nodes -key /etc/ssl/example.key -days 100000 -out /etc/ssl/example.cert -subj '/CN=Cloudy Health Checks Example by Madflojo'

CMD ["cloudy-healthchecks-example"]
