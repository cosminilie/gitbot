FROM golang:1.8 
MAINTAINER Cosmin Ilie <cosmin_ilie@live.com>

ENV GOPATH /go
ENV USER root


COPY . /go/src/github.com/cosminilie/gitbot

RUN cd /go/src/github.com/cosminilie/gitbot \
	&& go get -d -v \
	&& go build -o gitbot cmd/main.go \
	&& mv gitbot /go/bin \
	&& go test github.com/cosminilie/gitbot...
