FROM golang:latest

ADD . /go/src/github.com/madflojo/go-quick
WORKDIR /go/src/github.com/madflojo/go-quick/cmd/go-quick
RUN go install -v .

ENTRYPOINT ["../../docker-entrypoint.sh"]
