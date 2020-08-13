FROM golang:latest

ADD . /go/src/github.com/madflojo/healthchecks-example
WORKDIR /go/src/github.com/madflojo/healthchecks-example/cmd/healthchecks-example
RUN go install -v .

ENTRYPOINT ["../../docker-entrypoint.sh"]
