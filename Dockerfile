FROM golang:1.7.0

RUN apt-get update
COPY . /go/src/github.com/remind101/dockerdog
RUN go install github.com/remind101/dockerdog/...
